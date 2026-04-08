#!/usr/bin/env python3
"""
Model evaluation script for astrology AI.

This script evaluates trained models on astrology-specific tasks and metrics.
It computes perplexity, BLEU, ROUGE, and custom astrology accuracy metrics.
"""

import os
import json
import yaml
import torch
import logging
import numpy as np
from pathlib import Path
from typing import Dict, Any, List, Optional
from datetime import datetime
import matplotlib.pyplot as plt
import seaborn as sns

import transformers
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    pipeline,
    BitsAndBytesConfig
)
from peft import PeftModel
import datasets
from datasets import Dataset
from nltk.translate.bleu_score import sentence_bleu, SmoothingFunction
from rouge_score import rouge_scorer
import nltk

# Download required NLTK data
try:
    nltk.data.find('tokenizers/punkt')
except LookupError:
    nltk.download('punkt')

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class AstrologyModelEvaluator:
    """Handles model evaluation for astrology AI."""

    def __init__(self, config_path: str, model_path: Optional[str] = None):
        """Initialize with configuration."""
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)

        self.model_path = model_path or self.config['evaluation']['model_path']
        self.output_dir = Path("evaluation_results")
        self.output_dir.mkdir(exist_ok=True)

        # Create logs directory
        self.logs_dir = Path("logs")
        self.logs_dir.mkdir(exist_ok=True)

        # Setup device
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        logger.info(f"Using device: {self.device}")

    def load_model_and_tokenizer(self):
        """Load model and tokenizer."""
        logger.info(f"Loading model from: {self.model_path}")

        # Load tokenizer
        self.tokenizer = AutoTokenizer.from_pretrained(
            self.model_path,
            trust_remote_code=True
        )

        # Load model
        model_kwargs = {
            "trust_remote_code": True,
            "torch_dtype": torch.float16 if torch.cuda.is_available() else torch.float32,
        }

        # Check if this is a LoRA model
        adapter_config_path = Path(self.model_path) / "adapter_config.json"
        if adapter_config_path.exists():
            logger.info("Loading LoRA model...")
            # Load base model first
            with open(adapter_config_path, 'r') as f:
                adapter_config = json.load(f)
                base_model_name = adapter_config['base_model_name_or_path']

            base_model = AutoModelForCausalLM.from_pretrained(
                base_model_name,
                **model_kwargs
            )

            self.model = PeftModel.from_pretrained(base_model, self.model_path)
        else:
            self.model = AutoModelForCausalLM.from_pretrained(
                self.model_path,
                **model_kwargs
            )

        self.model.to(self.device)
        self.model.eval()

        logger.info(f"Model loaded with {self.model.num_parameters():,} parameters")

    def load_test_data(self) -> Dataset:
        """Load test data."""
        test_file = self.config['evaluation']['test_file']
        logger.info(f"Loading test data from: {test_file}")

        if not Path(test_file).exists():
            logger.warning(f"Test file not found: {test_file}, generating sample data...")
            return self._generate_sample_test_data()

        dataset = Dataset.from_json(test_file)
        logger.info(f"Loaded {len(dataset)} test examples")
        return dataset

    def _generate_sample_test_data(self) -> Dataset:
        """Generate sample test data."""
        sample_data = [
            {
                "instruction": "Write a daily horoscope for Aries.",
                "input": "",
                "output": "Dear Aries, today brings new opportunities. The stars align to bring success in your endeavors.",
                "category": "daily_horoscope",
                "signs": ["Aries"]
            },
            {
                "instruction": "Are Aries and Libra compatible?",
                "input": "",
                "output": "Aries and Libra have moderate compatibility. While they complement each other in some ways, they may face challenges in communication.",
                "category": "compatibility",
                "signs": ["Aries", "Libra"]
            },
            {
                "instruction": "What remedies should a Leo follow?",
                "input": "",
                "output": "For Leo, wear ruby gemstones, practice sun salutations, and maintain a positive attitude towards challenges.",
                "category": "remedies",
                "signs": ["Leo"]
            }
        ]

        return Dataset.from_list(sample_data)

    def generate_responses(self, dataset: Dataset) -> List[Dict[str, Any]]:
        """Generate responses for test prompts."""
        logger.info("Generating responses...")

        generation_config = self.config['evaluation']['generation']

        # Create text generation pipeline
        generator = pipeline(
            "text-generation",
            model=self.model,
            tokenizer=self.tokenizer,
            device=self.device,
            torch_dtype=torch.float16 if torch.cuda.is_available() else torch.float32,
        )

        results = []

        for i, example in enumerate(dataset):
            instruction = example['instruction']
            input_text = example['input']
            expected_output = example['output']

            # Format prompt
            if input_text.strip():
                prompt = f"### Instruction:\n{instruction}\n\n### Input:\n{input_text}\n\n### Response:\n"
            else:
                prompt = f"### Instruction:\n{instruction}\n\n### Response:\n"

            try:
                # Generate response
                outputs = generator(
                    prompt,
                    max_new_tokens=generation_config['max_new_tokens'],
                    temperature=generation_config['temperature'],
                    top_p=generation_config['top_p'],
                    top_k=generation_config['top_k'],
                    repetition_penalty=generation_config['repetition_penalty'],
                    do_sample=generation_config['do_sample'],
                    pad_token_id=generation_config['pad_token_id'],
                    eos_token_id=generation_config['eos_token_id'],
                    num_return_sequences=1,
                )

                generated_text = outputs[0]['generated_text']
                # Extract only the response part
                response_start = generated_text.find("### Response:\n")
                if response_start != -1:
                    generated_response = generated_text[response_start + len("### Response:\n"):].strip()
                else:
                    generated_response = generated_text.replace(prompt, "").strip()

                result = {
                    "index": i,
                    "instruction": instruction,
                    "input": input_text,
                    "expected_output": expected_output,
                    "generated_response": generated_response,
                    "category": example.get('category', 'general'),
                    "signs": example.get('signs', []),
                    "prompt": prompt
                }

                results.append(result)

                if (i + 1) % 10 == 0:
                    logger.info(f"Generated responses for {i + 1}/{len(dataset)} examples")

            except Exception as e:
                logger.error(f"Failed to generate response for example {i}: {e}")
                results.append({
                    "index": i,
                    "instruction": instruction,
                    "input": input_text,
                    "expected_output": expected_output,
                    "generated_response": "",
                    "category": example.get('category', 'general'),
                    "signs": example.get('signs', []),
                    "error": str(e)
                })

        return results

    def compute_metrics(self, results: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Compute evaluation metrics."""
        logger.info("Computing metrics...")

        metrics_config = self.config['evaluation']['metrics']
        enabled_metrics = [m['name'] for m in metrics_config if m['enabled']]

        scores = {}

        # Perplexity
        if 'perplexity' in enabled_metrics:
            scores['perplexity'] = self._compute_perplexity(results)

        # BLEU score
        if 'bleu' in enabled_metrics:
            scores['bleu'] = self._compute_bleu(results)

        # ROUGE scores
        if 'rouge' in enabled_metrics:
            rouge_scores = self._compute_rouge(results)
            scores.update(rouge_scores)

        # METEOR score
        if 'meteor' in enabled_metrics:
            scores['meteor'] = self._compute_meteor(results)

        # Astrology accuracy
        if 'astrology_accuracy' in enabled_metrics:
            scores['astrology_accuracy'] = self._compute_astrology_accuracy(results)

        # Overall scores
        scores['num_examples'] = len(results)
        scores['avg_response_length'] = np.mean([len(r['generated_response']) for r in results])

        return scores

    def _compute_perplexity(self, results: List[Dict[str, Any]]) -> float:
        """Compute perplexity on the test set."""
        logger.info("Computing perplexity...")

        total_loss = 0
        total_tokens = 0

        with torch.no_grad():
            for result in results:
                text = result['expected_output']
                inputs = self.tokenizer(text, return_tensors="pt").to(self.device)

                outputs = self.model(**inputs, labels=inputs['input_ids'])
                loss = outputs.loss.item()

                total_loss += loss * inputs['input_ids'].size(1)
                total_tokens += inputs['input_ids'].size(1)

        perplexity = np.exp(total_loss / total_tokens)
        return perplexity

    def _compute_bleu(self, results: List[Dict[str, Any]]) -> float:
        """Compute BLEU score."""
        logger.info("Computing BLEU score...")

        smoothing = SmoothingFunction().method4
        bleu_scores = []

        for result in results:
            reference = nltk.word_tokenize(result['expected_output'].lower())
            candidate = nltk.word_tokenize(result['generated_response'].lower())

            if reference and candidate:
                bleu = sentence_bleu([reference], candidate, smoothing_function=smoothing)
                bleu_scores.append(bleu)

        return np.mean(bleu_scores) if bleu_scores else 0.0

    def _compute_rouge(self, results: List[Dict[str, Any]]) -> Dict[str, float]:
        """Compute ROUGE scores."""
        logger.info("Computing ROUGE scores...")

        scorer = rouge_scorer.RougeScorer(['rouge1', 'rouge2', 'rougeL'], use_stemmer=True)

        rouge_scores = {'rouge1': [], 'rouge2': [], 'rougeL': []}

        for result in results:
            scores = scorer.score(result['expected_output'], result['generated_response'])

            for key in rouge_scores.keys():
                rouge_scores[key].append(scores[key].fmeasure)

        return {key: np.mean(values) for key, values in rouge_scores.items()}

    def _compute_meteor(self, results: List[Dict[str, Any]]) -> float:
        """Compute METEOR score."""
        logger.info("Computing METEOR score...")

        try:
            from nltk.translate.meteor_score import meteor_score
        except ImportError:
            logger.warning("METEOR score requires additional NLTK data. Skipping...")
            return 0.0

        meteor_scores = []

        for result in results:
            reference = nltk.word_tokenize(result['expected_output'].lower())
            candidate = nltk.word_tokenize(result['generated_response'].lower())

            if reference and candidate:
                meteor = meteor_score([reference], candidate)
                meteor_scores.append(meteor)

        return np.mean(meteor_scores) if meteor_scores else 0.0

    def _compute_astrology_accuracy(self, results: List[Dict[str, Any]]) -> float:
        """Compute astrology-specific accuracy."""
        logger.info("Computing astrology accuracy...")

        astrology_config = self.config['astrology_eval']
        accuracy_scores = []

        for result in results:
            response = result['generated_response'].lower()
            score = 0

            # Check for required astrology terms
            astrology_terms = astrology_config['quality_checks']['astrology_terms']
            found_terms = sum(1 for term in astrology_terms if term in response)
            required_terms = astrology_config['quality_checks']['astrology_terms_required']

            if found_terms >= required_terms:
                score += 0.5

            # Check content length
            min_length = astrology_config['quality_checks']['min_length']
            max_length = astrology_config['quality_checks']['max_length']

            if min_length <= len(response) <= max_length:
                score += 0.3

            # Check for positive sentiment (simple heuristic)
            if astrology_config['quality_checks']['sentiment_check']:
                positive_words = ['good', 'positive', 'fortunate', 'lucky', 'success', 'opportunity']
                found_positive = sum(1 for word in positive_words if word in response)
                min_positive = astrology_config['quality_checks']['min_positive_words']

                if found_positive >= min_positive:
                    score += 0.2

            accuracy_scores.append(score)

        return np.mean(accuracy_scores) if accuracy_scores else 0.0

    def save_results(self, results: List[Dict[str, Any]], metrics: Dict[str, Any]):
        """Save evaluation results."""
        logger.info("Saving results...")

        # Save detailed results
        detailed_file = self.output_dir / self.config['output']['detailed_results_file']
        with open(detailed_file, 'w') as f:
            json.dump(results, f, indent=2)

        # Save metrics
        results_file = self.output_dir / self.config['output']['results_file']
        with open(results_file, 'w') as f:
            json.dump(metrics, f, indent=2)

        # Save sample outputs
        sample_file = self.output_dir / self.config['output']['sample_outputs_file']
        with open(sample_file, 'w') as f:
            for i, result in enumerate(results[:10]):  # Save first 10 samples
                f.write(f"=== Sample {i+1} ===\n")
                f.write(f"Instruction: {result['instruction']}\n")
                f.write(f"Expected: {result['expected_output']}\n")
                f.write(f"Generated: {result['generated_response']}\n\n")

        logger.info(f"Results saved to {self.output_dir}")

    def create_plots(self, metrics: Dict[str, Any]):
        """Create evaluation plots."""
        logger.info("Creating plots...")

        plots_dir = Path(self.config['output']['plots_dir'])
        plots_dir.mkdir(exist_ok=True)

        # Create a simple bar plot of metrics
        metric_names = [k for k in metrics.keys() if isinstance(metrics[k], (int, float)) and k != 'num_examples']
        metric_values = [metrics[k] for k in metric_names]

        plt.figure(figsize=(10, 6))
        bars = plt.bar(metric_names, metric_values)
        plt.title('Model Evaluation Metrics')
        plt.ylabel('Score')
        plt.xticks(rotation=45)

        # Add value labels on bars
        for bar, value in zip(bars, metric_values):
            plt.text(bar.get_x() + bar.get_width()/2, bar.get_height(),
                    f'{value:.3f}', ha='center', va='bottom')

        plt.tight_layout()
        plt.savefig(plots_dir / 'evaluation_metrics.png', dpi=300, bbox_inches='tight')
        plt.close()

        logger.info(f"Plots saved to {plots_dir}")

    def run_evaluation(self):
        """Run the complete evaluation process."""
        logger.info("Starting evaluation...")

        try:
            # Load model and tokenizer
            self.load_model_and_tokenizer()

            # Load test data
            test_dataset = self.load_test_data()

            # Generate responses
            results = self.generate_responses(test_dataset)

            # Compute metrics
            metrics = self.compute_metrics(results)

            # Save results
            self.save_results(results, metrics)

            # Create plots
            self.create_plots(metrics)

            logger.info("Evaluation completed successfully!")
            logger.info(f"Results: {metrics}")

            return metrics

        except Exception as e:
            logger.error(f"Evaluation failed: {e}")
            raise

def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="Evaluate astrology AI model")
    parser.add_argument("--config", default="config/eval_config.yaml", help="Path to config file")
    parser.add_argument("--model-path", help="Path to trained model (overrides config)")
    args = parser.parse_args()

    evaluator = AstrologyModelEvaluator(args.config, args.model_path)
    metrics = evaluator.run_evaluation()

    print("\nEvaluation Results:")
    for key, value in metrics.items():
        if isinstance(value, float):
            print(".4f")
        else:
            print(f"{key}: {value}")

if __name__ == "__main__":
    main()