#!/bin/bash

# Comprehensive Astrology AI Training Guide
# This script provides step-by-step instructions for training the astrology AI model

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DATA_DIR="$PROJECT_ROOT/data"
MODELS_DIR="$PROJECT_ROOT/models"
LOGS_DIR="$PROJECT_ROOT/logs"

# Create necessary directories
mkdir -p "$DATA_DIR" "$MODELS_DIR" "$LOGS_DIR"

# Run from repo root so config paths resolve correctly
cd "$PROJECT_ROOT"

# Function to print colored output
print_header() {
    echo -e "${PURPLE}================================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}================================================${NC}"
}

print_step() {
    echo -e "${CYAN}[STEP $1]${NC} $2"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_header "CHECKING PREREQUISITES"

    local missing_deps=()

    if ! command -v python3 &> /dev/null; then
        missing_deps+=("python3")
    fi

    if ! command -v pip3 &> /dev/null; then
        missing_deps+=("pip3")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        print_info "Please install missing dependencies and try again."
        exit 1
    fi

    print_success "All prerequisites met."
}

# Function to install Python dependencies
install_dependencies() {
    print_header "INSTALLING PYTHON DEPENDENCIES"

    if [ ! -f "$SCRIPT_DIR/requirements-train.txt" ]; then
        print_error "requirements-train.txt not found in $SCRIPT_DIR"
        exit 1
    fi

    print_info "Installing Python packages..."
    pip3 install -r "$SCRIPT_DIR/requirements-train.txt"

    print_success "Python dependencies installed."
}

# Function to generate expanded training data
generate_expanded_data() {
    print_header "GENERATING EXPANDED TRAINING DATA"

    print_info "Creating comprehensive astrology training dataset..."
    python3 "$SCRIPT_DIR/generate_expanded_data.py"

    print_success "Expanded training data generated."
}

# Function to prepare training data
prepare_data() {
    print_header "PREPARING TRAINING DATA"

    print_info "Processing and splitting training data..."
    python3 "$SCRIPT_DIR/prepare_data.py" --config "$SCRIPT_DIR/config/data_config.yaml"

    print_success "Training data prepared."
}

# Function to train model
train_model() {
    print_header "TRAINING THE ASTROLOGY AI MODEL"

    print_info "Starting model training with LoRA fine-tuning..."
    print_info "This may take several minutes to hours depending on your hardware."

    # Check available memory/hardware
    python3 -c "
import torch
print(f'CUDA available: {torch.cuda.is_available()}')
if torch.cuda.is_available():
    print(f'GPU count: {torch.cuda.device_count()}')
    for i in range(torch.cuda.device_count()):
        print(f'GPU {i}: {torch.cuda.get_device_name(i)}')
else:
    print('Running on CPU - training will be slower')
"

    # Start training
    python3 "$SCRIPT_DIR/train_model.py" --config "$SCRIPT_DIR/config/train_config.yaml"

    print_success "Training completed!"
}

# Function to evaluate model
evaluate_model() {
    print_header "EVALUATING TRAINED MODEL"

    local model_path="${1:-models/astrology-ai/final_model}"

    if [ ! -d "$model_path" ]; then
        print_error "Model directory not found: $model_path"
        print_info "Please train the model first or specify correct path."
        exit 1
    fi

    print_info "Evaluating trained model..."
    python3 "$SCRIPT_DIR/evaluate_model.py" --config "$SCRIPT_DIR/config/eval_config.yaml" --model-path "$model_path"

    print_success "Evaluation completed."
}

# Function to create Ollama model
create_ollama_model() {
    print_header "CREATING OLLAMA MODEL"

    local model_path="${1:-models/astrology-ai/final_model}"

    if [ ! -d "$model_path" ]; then
        print_error "Model directory not found: $model_path"
        exit 1
    fi

    print_info "Creating Ollama model from trained checkpoint..."
    python3 "$SCRIPT_DIR/create_ollama_model.py" --config "$SCRIPT_DIR/config/train_config.yaml" --model-path "$model_path"

    print_success "Ollama model created."
}

# Function to run full training pipeline
full_training_pipeline() {
    print_header "ASTROLOGY AI TRAINING PIPELINE"
    print_info "This will run the complete training process from data preparation to model deployment."

    echo
    print_step "1" "Checking prerequisites..."
    check_prerequisites

    echo
    print_step "2" "Installing Python dependencies..."
    install_dependencies

    echo
    print_step "3" "Generating expanded training data..."
    generate_expanded_data

    echo
    print_step "4" "Preparing training data..."
    prepare_data

    echo
    print_step "5" "Training the model..."
    train_model

    echo
    print_step "6" "Evaluating the model..."
    evaluate_model

    echo
    print_step "7" "Creating Ollama model..."
    create_ollama_model

    print_header "TRAINING PIPELINE COMPLETED"
    print_success "Your Astrology AI model is ready!"
    print_info "Model saved to: models/astrology-ai/final_model"
    print_info "Ollama model created for local inference"
}

# Function to show usage
show_usage() {
    print_header "ASTROLOGY AI TRAINING GUIDE"

    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  full_pipeline     - Run complete training pipeline (recommended)"
    echo "  check_prereqs     - Check system prerequisites"
    echo "  install_deps      - Install Python dependencies"
    echo "  generate_data     - Generate expanded training data"
    echo "  prepare_data      - Prepare and split training data"
    echo "  train             - Train the model"
    echo "  evaluate          - Evaluate trained model"
    echo "  create_ollama     - Create Ollama model for inference"
    echo "  help              - Show this help message"
    echo
    echo "Examples:"
    echo "  $0 full_pipeline              # Complete training"
    echo "  $0 train                      # Just training step"
    echo "  $0 evaluate models/my_model   # Evaluate specific model"
    echo
    print_info "For first-time setup, use: $0 full_pipeline"
}

# Main script logic
case "${1:-help}" in
    "full_pipeline")
        full_training_pipeline
        ;;
    "check_prereqs")
        check_prerequisites
        ;;
    "install_deps")
        install_dependencies
        ;;
    "generate_data")
        generate_expanded_data
        ;;
    "prepare_data")
        prepare_data
        ;;
    "train")
        train_model
        ;;
    "evaluate")
        evaluate_model "$2"
        ;;
    "create_ollama")
        create_ollama_model "$2"
        ;;
    "help"|*)
        show_usage
        ;;
esac