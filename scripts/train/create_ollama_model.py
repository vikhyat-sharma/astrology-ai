#!/usr/bin/env python3
"""
Ollama model creation script for astrology AI.

This script converts a trained model to Ollama format and creates a Modelfile
for easy deployment and inference.
"""

import os
import json
import yaml
import shutil
import logging
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class OllamaModelCreator:
    """Handles creation of Ollama models from trained checkpoints."""

    def __init__(self, config_path: str, model_path: str):
        """Initialize with configuration and model path."""
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)

        self.model_path = Path(model_path)
        self.output_dir = Path("ollama_models")
        self.output_dir.mkdir(exist_ok=True)

        # Create logs directory
        self.logs_dir = Path("logs")
        self.logs_dir.mkdir(exist_ok=True)

    def create_modelfile(self) -> str:
        """Create Ollama Modelfile for the astrology model."""
        logger.info("Creating Ollama Modelfile...")

        # Load training config to get base model info
        training_config_path = self.model_path / "training_config.json"
        if training_config_path.exists():
            with open(training_config_path, 'r') as f:
                training_config = json.load(f)
                base_model = training_config['model']['base_model']
        else:
            # Fallback to config
            base_model = self.config['model']['base_model']

        # Create Modelfile content
        modelfile_content = f"""FROM {base_model}

# Astrology AI Model
PARAMETER temperature 0.7
PARAMETER top_p 0.9
PARAMETER top_k 50
PARAMETER repeat_penalty 1.1
PARAMETER num_ctx 2048

SYSTEM \"\"\"You are an expert astrologer providing accurate, helpful, and positive astrology readings. You specialize in horoscopes, compatibility analysis, and astrological remedies. Always provide detailed, personalized insights based on zodiac signs, planetary positions, and traditional astrological wisdom.\"\"\"

# Template for instruction-response format
TEMPLATE \"\"\"{{{{ if .System }}}}{{{{.System}}}}
{{{{ end }}}}{{{{.Prompt}}}}
{{{{ if .Response }}}}{{{{.Response}}}}
{{{{ end }}}}\"\"\"

# Additional parameters for fine-tuned model
PARAMETER stop "### Instruction:"
PARAMETER stop "### Input:"
PARAMETER stop "### Response:"
"""

        # Save Modelfile
        modelfile_path = self.output_dir / "Modelfile"
        with open(modelfile_path, 'w') as f:
            f.write(modelfile_content)

        logger.info(f"Modelfile created at {modelfile_path}")
        return str(modelfile_path)

    def create_model_metadata(self) -> Dict[str, Any]:
        """Create metadata for the Ollama model."""
        logger.info("Creating model metadata...")

        metadata = {
            "name": "astrology-ai",
            "version": datetime.now().strftime("%Y%m%d-%H%M%S"),
            "description": "Fine-tuned language model for astrology readings, horoscopes, and compatibility analysis",
            "author": "Astrology AI Training Pipeline",
            "license": "Apache 2.0",
            "base_model": self.config['model']['base_model'],
            "training_config": {
                "lora_r": self.config['lora']['r'],
                "lora_alpha": self.config['lora']['lora_alpha'],
                "epochs": self.config['training']['num_train_epochs'],
                "learning_rate": self.config['training']['learning_rate']
            },
            "capabilities": [
                "Daily horoscopes",
                "Weekly horoscopes",
                "Monthly horoscopes",
                "Compatibility analysis",
                "Astrological remedies",
                "Birth chart insights"
            ],
            "parameters": {
                "temperature": 0.7,
                "top_p": 0.9,
                "top_k": 50,
                "repeat_penalty": 1.1,
                "context_length": 2048
            },
            "created_at": datetime.now().isoformat()
        }

        # Save metadata
        metadata_path = self.output_dir / "metadata.json"
        with open(metadata_path, 'w') as f:
            json.dump(metadata, f, indent=2)

        logger.info(f"Metadata saved to {metadata_path}")
        return metadata

    def copy_model_files(self):
        """Copy necessary model files for Ollama."""
        logger.info("Copying model files...")

        # Files to copy from the trained model directory
        files_to_copy = [
            "adapter_config.json",
            "adapter_model.bin",
            "adapter_model.safetensors",
            "tokenizer.json",
            "tokenizer_config.json",
            "special_tokens_map.json",
            "vocab.json",
            "merges.txt",
            "added_tokens.json"
        ]

        copied_files = []
        for file_name in files_to_copy:
            src_path = self.model_path / file_name
            if src_path.exists():
                dst_path = self.output_dir / file_name
                shutil.copy2(src_path, dst_path)
                copied_files.append(file_name)
                logger.debug(f"Copied {file_name}")

        logger.info(f"Copied {len(copied_files)} model files: {', '.join(copied_files)}")

        # Check for safetensors files (preferred format)
        safetensors_files = list(self.model_path.glob("*.safetensors"))
        if safetensors_files:
            logger.info(f"Found {len(safetensors_files)} safetensors files")
            for safetensors_file in safetensors_files:
                dst_path = self.output_dir / safetensors_file.name
                shutil.copy2(safetensors_file, dst_path)
                logger.debug(f"Copied {safetensors_file.name}")

    def create_readme(self, metadata: Dict[str, Any]):
        """Create README for the Ollama model."""
        logger.info("Creating README...")

        readme_content = f"""# Astrology AI Ollama Model

{metadata['description']}

## Model Information

- **Name**: {metadata['name']}
- **Version**: {metadata['version']}
- **Base Model**: {metadata['base_model']}
- **Created**: {metadata['created_at']}

## Capabilities

This model specializes in:
{chr(10).join(f"- {capability}" for capability in metadata['capabilities'])}

## Usage

### With Ollama CLI

1. Create the model:
```bash
ollama create astrology-ai -f Modelfile
```

2. Run the model:
```bash
ollama run astrology-ai
```

### Example Prompts

#### Daily Horoscope
```
Write a daily horoscope for Aries for today.
```

#### Compatibility Analysis
```
Are Aries and Libra compatible? What are their strengths and challenges?
```

#### Astrological Remedies
```
What remedies should a Leo follow to improve their financial situation?
```

## Parameters

- **Temperature**: {metadata['parameters']['temperature']}
- **Top P**: {metadata['parameters']['top_p']}
- **Top K**: {metadata['parameters']['top_k']}
- **Repeat Penalty**: {metadata['parameters']['repeat_penalty']}
- **Context Length**: {metadata['parameters']['context_length']}

## Training Details

- LoRA Rank: {metadata['training_config']['lora_r']}
- LoRA Alpha: {metadata['training_config']['lora_alpha']}
- Training Epochs: {metadata['training_config']['epochs']}
- Learning Rate: {metadata['training_config']['learning_rate']}

## License

{metadata['license']}

## Author

{metadata['author']}
"""

        # Save README
        readme_path = self.output_dir / "README.md"
        with open(readme_path, 'w') as f:
            f.write(readme_content)

        logger.info(f"README created at {readme_path}")

    def create_install_script(self):
        """Create installation script for easy setup."""
        logger.info("Creating installation script...")

        install_script = """#!/bin/bash
# Astrology AI Model Installation Script

set -e

MODEL_NAME="astrology-ai"
MODEL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing Astrology AI model for Ollama..."

# Check if ollama is installed
if ! command -v ollama &> /dev/null; then
    echo "Error: Ollama is not installed. Please install Ollama first:"
    echo "  curl -fsSL https://ollama.ai/install.sh | sh"
    exit 1
fi

# Create the model
echo "Creating Ollama model..."
cd "$MODEL_DIR"
ollama create "$MODEL_NAME" -f Modelfile

# Verify installation
echo "Verifying installation..."
if ollama list | grep -q "$MODEL_NAME"; then
    echo "✅ Astrology AI model installed successfully!"
    echo ""
    echo "To use the model:"
    echo "  ollama run $MODEL_NAME"
    echo ""
    echo "Example usage:"
    echo "  ollama run $MODEL_NAME"
    echo "  >>> Write a daily horoscope for Taurus"
else
    echo "❌ Model installation failed"
    exit 1
fi
"""

        # Save install script
        install_path = self.output_dir / "install.sh"
        with open(install_path, 'w') as f:
            f.write(install_script)

        # Make executable
        os.chmod(install_path, 0o755)

        logger.info(f"Installation script created at {install_path}")

    def create_archive(self):
        """Create a compressed archive of the model."""
        logger.info("Creating model archive...")

        import tarfile

        archive_name = f"astrology-ai-model-{datetime.now().strftime('%Y%m%d-%H%M%S')}.tar.gz"
        archive_path = self.output_dir.parent / archive_name

        with tarfile.open(archive_path, "w:gz") as tar:
            tar.add(str(self.output_dir), arcname="astrology-ai-model")

        logger.info(f"Model archive created: {archive_path}")
        return str(archive_path)

    def run(self):
        """Run the complete Ollama model creation process."""
        logger.info("Starting Ollama model creation...")

        try:
            # Create Modelfile
            modelfile_path = self.create_modelfile()

            # Create metadata
            metadata = self.create_model_metadata()

            # Copy model files
            self.copy_model_files()

            # Create README
            self.create_readme(metadata)

            # Create installation script
            self.create_install_script()

            # Create archive
            archive_path = self.create_archive()

            logger.info("Ollama model creation completed successfully!")
            logger.info(f"Model files are ready in: {self.output_dir}")
            logger.info(f"To install: cd {self.output_dir} && ./install.sh")

            return {
                "model_dir": str(self.output_dir),
                "modelfile": modelfile_path,
                "archive": archive_path,
                "metadata": metadata
            }

        except Exception as e:
            logger.error(f"Ollama model creation failed: {e}")
            raise

def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="Create Ollama model from trained checkpoint")
    parser.add_argument("--config", default="config/train_config.yaml", help="Path to training config file")
    parser.add_argument("--model-path", required=True, help="Path to trained model directory")
    parser.add_argument("--archive", action="store_true", help="Create compressed archive")
    args = parser.parse_args()

    creator = OllamaModelCreator(args.config, args.model_path)
    result = creator.run()

    print("\nOllama Model Creation Complete!")
    print(f"Model directory: {result['model_dir']}")
    print(f"Modelfile: {result['modelfile']}")
    if args.archive:
        print(f"Archive: {result['archive']}")

    print("\nTo install the model:")
    print(f"  cd {result['model_dir']}")
    print("  ./install.sh")

if __name__ == "__main__":
    main()