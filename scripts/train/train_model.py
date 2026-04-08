#!/usr/bin/env python3
"""
Model training script for astrology AI.

This script fine-tunes a language model on astrology-specific data using LoRA.
It supports distributed training and various optimization techniques.
"""

import os
import json
import yaml
import torch
import logging
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

import transformers
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    TrainingArguments,
    Trainer,
    DataCollatorForLanguageModeling,
    EarlyStoppingCallback
)
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
import datasets
from datasets import Dataset

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class AstrologyModelTrainer:
    """Handles model training for astrology AI."""

    def __init__(self, config_path: str):
        """Initialize with configuration."""
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)

        self.model_name = self.config['model']['base_model']
        self.output_dir = Path(self.config['training']['output_dir'])
        self.output_dir.mkdir(parents=True, exist_ok=True)

        # Create logs directory
        self.logs_dir = Path("logs")
        self.logs_dir.mkdir(exist_ok=True)

        # Setup device
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        logger.info(f"Using device: {self.device}")

        if torch.cuda.is_available():
            logger.info(f"GPU: {torch.cuda.get_device_name(0)}")
            logger.info(f"CUDA version: {torch.version.cuda}")

    def load_data(self) -> Dict[str, Dataset]:
        """Load processed training data."""
        logger.info("Loading training data...")

        data_dir = Path(self.config['data']['processed_data_dir'])

        datasets_dict = {}
        for split in ['train', 'validation']:
            file_path = data_dir / f"{split}.jsonl"
            if file_path.exists():
                dataset = Dataset.from_json(str(file_path))
                datasets_dict[split] = dataset
                logger.info(f"Loaded {len(dataset)} {split} examples")
            else:
                raise FileNotFoundError(f"Data file not found: {file_path}")

        return datasets_dict

    def load_model_and_tokenizer(self):
        """Load base model and tokenizer."""
        logger.info(f"Loading model: {self.model_name}")

        # Load tokenizer
        self.tokenizer = AutoTokenizer.from_pretrained(
            self.model_name,
            trust_remote_code=True
        )

        # Add padding token if not exists
        if self.tokenizer.pad_token is None:
            self.tokenizer.pad_token = self.tokenizer.eos_token

        # Load model with quantization if specified
        model_kwargs = {
            "trust_remote_code": True,
            "torch_dtype": torch.float16 if torch.cuda.is_available() else torch.float32,
        }

        if self.config['model'].get('load_in_8bit', False):
            model_kwargs["load_in_8bit"] = True
        elif self.config['model'].get('load_in_4bit', False):
            model_kwargs["load_in_4bit"] = True

        self.model = AutoModelForCausalLM.from_pretrained(
            self.model_name,
            **model_kwargs
        )

        logger.info(f"Model loaded with {self.model.num_parameters():,} parameters")

    def setup_lora(self):
        """Setup LoRA configuration."""
        logger.info("Setting up LoRA...")

        lora_config = LoraConfig(
            r=self.config['lora']['r'],
            lora_alpha=self.config['lora']['lora_alpha'],
            target_modules=self.config['lora']['target_modules'],
            lora_dropout=self.config['lora']['lora_dropout'],
            bias=self.config['lora']['bias'],
            task_type="CAUSAL_LM"
        )

        # Prepare model for training
        if self.config['model'].get('load_in_8bit', False) or self.config['model'].get('load_in_4bit', False):
            self.model = prepare_model_for_kbit_training(self.model)

        self.model = get_peft_model(self.model, lora_config)

        # Print trainable parameters
        trainable_params = sum(p.numel() for p in self.model.parameters() if p.requires_grad)
        total_params = sum(p.numel() for p in self.model.parameters())
        logger.info(f"Trainable parameters: {trainable_params:,} ({100 * trainable_params / total_params:.2f}%)")

    def preprocess_function(self, examples):
        """Preprocess data for training."""
        # Format instruction-response pairs
        formatted_texts = []
        for instruction, input_text, output_text in zip(
            examples['instruction'],
            examples['input'],
            examples['output']
        ):
            if input_text.strip():
                text = f"### Instruction:\n{instruction}\n\n### Input:\n{input_text}\n\n### Response:\n{output_text}"
            else:
                text = f"### Instruction:\n{instruction}\n\n### Response:\n{output_text}"

            formatted_texts.append(text)

        # Tokenize
        tokenized = self.tokenizer(
            formatted_texts,
            truncation=True,
            padding=False,
            max_length=self.config['data']['max_seq_length'],
            return_tensors="pt"
        )

        return tokenized

    def setup_training_args(self) -> TrainingArguments:
        """Setup training arguments."""
        training_config = self.config['training']

        # Generate run name
        run_name = f"astrology-ai-{datetime.now().strftime('%Y%m%d-%H%M%S')}"

        args = TrainingArguments(
            output_dir=str(self.output_dir),
            run_name=run_name,

            # Training hyperparameters
            num_train_epochs=training_config['num_train_epochs'],
            per_device_train_batch_size=training_config['per_device_train_batch_size'],
            per_device_eval_batch_size=training_config['per_device_eval_batch_size'],
            gradient_accumulation_steps=training_config['gradient_accumulation_steps'],
            learning_rate=training_config['learning_rate'],
            weight_decay=training_config['weight_decay'],
            warmup_steps=training_config['warmup_steps'],
            lr_scheduler_type=training_config['lr_scheduler_type'],

            # Optimization
            fp16=training_config.get('fp16', True),
            bf16=training_config.get('bf16', False),
            gradient_checkpointing=training_config.get('gradient_checkpointing', True),

            # Evaluation and saving
            evaluation_strategy="steps",
            eval_steps=training_config['eval_steps'],
            save_steps=training_config['save_steps'],
            save_total_limit=training_config['save_total_limit'],
            load_best_model_at_end=True,
            metric_for_best_model="eval_loss",

            # Logging
            logging_steps=training_config['logging_steps'],
            report_to=training_config.get('report_to', []),

            # Other
            dataloader_num_workers=training_config.get('dataloader_num_workers', 0),
            remove_unused_columns=False,
        )

        return args

    def train(self):
        """Run the training process."""
        logger.info("Starting training...")

        try:
            # Load data
            datasets_dict = self.load_data()

            # Load model and tokenizer
            self.load_model_and_tokenizer()

            # Setup LoRA
            self.setup_lora()

            # Preprocess data
            logger.info("Preprocessing data...")
            tokenized_datasets = datasets_dict.copy()
            for split_name, dataset in datasets_dict.items():
                tokenized_datasets[split_name] = dataset.map(
                    self.preprocess_function,
                    batched=True,
                    remove_columns=dataset.column_names,
                    desc=f"Preprocessing {split_name} data"
                )

            # Setup data collator
            data_collator = DataCollatorForLanguageModeling(
                tokenizer=self.tokenizer,
                mlm=False
            )

            # Setup training arguments
            training_args = self.setup_training_args()

            # Setup callbacks
            callbacks = []
            if self.config['training'].get('early_stopping', False):
                callbacks.append(EarlyStoppingCallback(
                    early_stopping_patience=self.config['training']['early_stopping_patience']
                ))

            # Initialize trainer
            trainer = Trainer(
                model=self.model,
                args=training_args,
                train_dataset=tokenized_datasets['train'],
                eval_dataset=tokenized_datasets['validation'],
                data_collator=data_collator,
                callbacks=callbacks,
            )

            # Start training
            logger.info("Training started...")
            trainer.train()

            # Save the final model
            final_model_path = self.output_dir / "final_model"
            trainer.save_model(str(final_model_path))
            self.tokenizer.save_pretrained(str(final_model_path))

            logger.info(f"Training completed! Model saved to {final_model_path}")

            # Save training config
            with open(self.output_dir / "training_config.json", 'w') as f:
                json.dump(self.config, f, indent=2)

            return str(final_model_path)

        except Exception as e:
            logger.error(f"Training failed: {e}")
            raise

def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="Train astrology AI model")
    parser.add_argument("--config", default="config/train_config.yaml", help="Path to config file")
    args = parser.parse_args()

    trainer = AstrologyModelTrainer(args.config)
    model_path = trainer.train()
    print(f"Training completed. Model saved to: {model_path}")

if __name__ == "__main__":
    main()