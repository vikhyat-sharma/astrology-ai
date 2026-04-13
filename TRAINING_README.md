# Astrology AI Training Guide

This guide provides comprehensive instructions for training the Astrology AI model, which generates personalized horoscopes, compatibility analysis, and Vedic astrology insights.

## 📊 Dataset Overview

The training dataset now includes **145 high-quality examples** covering:

- **Daily Horoscopes**: 31 examples across all 12 zodiac signs
- **Weekly Horoscopes**: 11 examples for comprehensive weekly guidance
- **Monthly Horoscopes**: 9 examples for monthly life guidance
- **Compatibility Analysis**: 13 examples for sign pair compatibility
- **Remedial Measures**: 12 examples of astrological remedies for each sign
- **Planetary Analysis**: 9 examples explaining planetary significance
- **House Analysis**: 12 examples of astrological house meanings
- **Yoga Identification**: 5 examples of planetary combinations
- **Birth Chart Readings**: 4 examples of chart interpretations
- **Dasha Predictions**: 8 examples of timing system predictions

## 🚀 Quick Start

### Option 1: Complete Automated Pipeline (Recommended)

```bash
# Run the complete training pipeline
./scripts/train/train_guide.sh full_pipeline
```

This will:
1. Check prerequisites
2. Install dependencies
3. Generate expanded training data
4. Prepare and split data
5. Train the model
6. Evaluate performance
7. Create Ollama model

### Option 2: Step-by-Step Training

```bash
# 1. Check prerequisites
./scripts/train/train_guide.sh check_prereqs

# 2. Install dependencies
./scripts/train/train_guide.sh install_deps

# 3. Generate training data
./scripts/train/train_guide.sh generate_data

# 4. Prepare data
./scripts/train/train_guide.sh prepare_data

# 5. Train model
./scripts/train/train_guide.sh train

# 6. Evaluate model
./scripts/train/train_guide.sh evaluate

# 7. Create Ollama model
./scripts/train/train_guide.sh create_ollama
```

## 📋 Prerequisites

### System Requirements
- **Python 3.8+**
- **pip** package manager
- **4GB+ RAM** (8GB recommended)
- **GPU optional** but recommended for faster training

### Python Dependencies
The training automatically installs all required packages:
- `torch` - PyTorch for deep learning
- `transformers` - Hugging Face transformers
- `peft` - Parameter-Efficient Fine-Tuning (LoRA)
- `datasets` - Hugging Face datasets
- `accelerate` - Training acceleration
- `wandb` - Experiment tracking (optional)

## 🏗️ Training Configuration

### Model Architecture
- **Base Model**: DialoGPT-medium (354M parameters)
- **Fine-tuning**: LoRA (Low-Rank Adaptation)
- **Trainable Parameters**: ~7.1M (2% of total)
- **Memory Efficient**: Suitable for CPU/GPU training

### Training Hyperparameters
```yaml
# Key settings in train_config.yaml
num_train_epochs: 3          # Training epochs
max_steps: 500              # Maximum training steps
learning_rate: 0.00002      # Conservative learning rate
batch_size: 2               # Per-device batch size
gradient_accumulation: 4     # Effective batch size = 8
warmup_steps: 50            # Learning rate warmup
```

### Data Configuration
- **Training**: 116 examples (80%)
- **Validation**: 14 examples (10%)
- **Test**: 15 examples (10%)
- **Max Length**: 512 tokens
- **Format**: Instruction-response pairs

## 🎯 Training Process

### Phase 1: Data Preparation
```bash
python3 scripts/train/generate_expanded_data.py
python3 scripts/train/prepare_data.py --config scripts/train/config/data_config.yaml
```

**Output**: Processed JSONL files in `data/processed/`

### Phase 2: Model Training
```bash
python3 scripts/train/train_model.py --config scripts/train/config/train_config.yaml
```

**Features**:
- LoRA fine-tuning for memory efficiency
- Gradient checkpointing
- Mixed precision training (if GPU available)
- Automatic checkpoint saving
- Training metrics logging

**Expected Output**:
- Loss: Decreasing from ~7.0 to ~3.0-4.0
- Training time: 10-30 minutes on GPU, 1-2 hours on CPU
- Model saved to: `models/astrology-ai/final_model/`

### Phase 3: Evaluation
```bash
python3 scripts/train/evaluate_model.py --config scripts/train/config/eval_config.yaml --model-path models/astrology-ai/final_model
```

**Metrics**:
- Perplexity
- BLEU score
- ROUGE scores
- Human evaluation readiness

### Phase 4: Deployment
```bash
python3 scripts/train/create_ollama_model.py --config scripts/train/config/train_config.yaml --model-path models/astrology-ai/final_model
```

**Output**: Ollama-compatible model for local inference

## 🔧 Customization Options

### Modifying Training Data
Add new examples to `data/raw/expanded_astrology_training.jsonl` following this format:
```json
{
  "instruction": "Write a daily horoscope for Aries.",
  "input": "",
  "output": "Dear Aries, today brings opportunities for...",
  "category": "daily_horoscope",
  "signs": ["Aries"],
  "metadata": {
    "input_length": 34,
    "output_length": 443,
    "created_at": "2026-04-11T14:50:14.552037"
  }
}
```

### Adjusting Training Parameters
Edit `scripts/train/config/train_config.yaml`:
- Increase `num_train_epochs` for better learning
- Adjust `learning_rate` if training is unstable
- Modify `lora.r` for different adaptation strength
- Change `max_steps` for longer/shorter training

### Hardware Optimization
For GPU training:
```yaml
# In train_config.yaml
training:
  fp16: true  # Enable mixed precision
  gradient_checkpointing: true  # Memory efficiency
```

## 📈 Expected Results

### Training Metrics
- **Initial Loss**: ~7.0-8.0
- **Final Loss**: ~3.0-4.0
- **Convergence**: Within 200-500 steps
- **Validation Loss**: Should decrease steadily

### Model Capabilities
After training, the model can generate:
- Personalized daily/weekly/monthly horoscopes
- Zodiac compatibility analysis
- Astrological remedies and guidance
- Vedic astrology explanations
- Birth chart interpretations

### Performance Benchmarks
- **Response Quality**: Coherent, astrology-specific content
- **Factuality**: Accurate zodiac and planetary information
- **Creativity**: Varied responses for similar queries
- **Consistency**: Maintains astrological context

## 🐛 Troubleshooting

### Common Issues

**Out of Memory Errors**:
```bash
# Reduce batch size
per_device_train_batch_size: 1
gradient_accumulation_steps: 8
```

**Slow Training**:
- Use GPU if available
- Reduce `max_steps`
- Enable gradient checkpointing

**Poor Model Performance**:
- Increase training data
- Adjust learning rate
- Train for more epochs
- Check data quality

**Import Errors**:
```bash
pip3 install --upgrade -r scripts/train/requirements-train.txt
```

### Getting Help
- Check training logs in `logs/training.log`
- Verify data format in `data/processed/`
- Test model loading: `python3 -c "from transformers import AutoModelForCausalLM; print('OK')"`

## 🚀 Next Steps

### Inference
After training, use the model for generation:
```python
from transformers import AutoTokenizer, AutoModelForCausalLM
import torch

model_path = "models/astrology-ai/final_model"
tokenizer = AutoTokenizer.from_pretrained(model_path)
model = AutoModelForCausalLM.from_pretrained(model_path)

def generate_horoscope(sign, period="daily"):
    prompt = f"Write a {period} horoscope for {sign}."
    inputs = tokenizer(prompt, return_tensors="pt")

    with torch.no_grad():
        outputs = model.generate(
            **inputs,
            max_new_tokens=200,
            temperature=0.7,
            top_p=0.9,
            do_sample=True
        )

    return tokenizer.decode(outputs[0], skip_special_tokens=True)
```

### Model Improvement
- **Add More Data**: Collect real horoscopes and astrology content
- **Fine-tune Further**: Train on domain-specific data
- **Experiment**: Try different base models (Llama, GPT-J)
- **Evaluate**: Use human evaluation for quality assessment

### Production Deployment
- Integrate with web API
- Add user personalization
- Implement caching for common queries
- Monitor performance and update regularly

---

**Happy Training!** 🌟🔮 The Astrology AI model will help bring astrological insights to users worldwide.