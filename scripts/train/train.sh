#!/bin/bash

# AI Training Scripts for Astrology AI
# This script provides utilities for training and fine-tuning AI models for astrology content generation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DATA_DIR="$PROJECT_ROOT/data"
MODELS_DIR="$PROJECT_ROOT/models"
LOGS_DIR="$PROJECT_ROOT/logs"

# Create necessary directories
mkdir -p "$DATA_DIR" "$MODELS_DIR" "$LOGS_DIR"

# Function to print colored output
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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."

    local missing_deps=()

    if ! command_exists python3; then
        missing_deps+=("python3")
    fi

    if ! command_exists pip3; then
        missing_deps+=("pip3")
    fi

    if ! command_exists ollama; then
        missing_deps+=("ollama")
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
    print_info "Installing Python dependencies..."

    if [ ! -f "$SCRIPT_DIR/requirements-train.txt" ]; then
        print_error "requirements-train.txt not found in $SCRIPT_DIR"
        exit 1
    fi

    pip3 install -r "$SCRIPT_DIR/requirements-train.txt"

    print_success "Python dependencies installed."
}

# Function to prepare training data
prepare_data() {
    print_info "Preparing training data..."

    python3 "$SCRIPT_DIR/prepare_data.py"

    print_success "Training data prepared."
}

# Function to train model
train_model() {
    print_info "Starting model training..."

    python3 "$SCRIPT_DIR/train_model.py"

    print_success "Training completed."
}

# Function to evaluate model
evaluate_model() {
    local model_path="${1:-models/final_model}"

    print_info "Evaluating trained model..."

    if [ ! -d "$model_path" ]; then
        print_error "Model directory not found: $model_path"
        echo "Usage: $0 evaluate_model [model_path]"
        exit 1
    fi

    python3 "$SCRIPT_DIR/evaluate_model.py" --model-path "$model_path"

    print_success "Evaluation completed."
}

# Function to create Ollama model
create_ollama_model() {
    local model_path="${1:-models/final_model}"

    print_info "Creating Ollama model from trained checkpoint..."

    if [ ! -d "$model_path" ]; then
        print_error "Model directory not found: $model_path"
        echo "Usage: $0 create_ollama_model [model_path]"
        exit 1
    fi

    python3 "$SCRIPT_DIR/create_ollama_model.py" --model-path "$model_path"

    print_success "Ollama model created."
}

# Function to run full training pipeline
full_training_pipeline() {
    print_info "Starting full training pipeline..."

    check_prerequisites
    install_dependencies
    prepare_data
    train_model
    evaluate_model
    create_ollama_model

    print_success "Full training pipeline completed!"
    print_info "You can now use the model with: ollama run astrology-ai"
}

# Function to show usage
show_usage() {
    cat << EOF
AI Training Scripts for Astrology AI

USAGE:
    $0 <command> [options]

COMMANDS:
    setup                    Setup training environment and install dependencies
    prepare-data             Prepare and preprocess training data
    train                    Train a new model using prepared data
    evaluate [model_path]    Evaluate a trained model (default: models/final_model)
    create-ollama [model_path] Create Ollama model from trained checkpoint
    full-pipeline            Run complete training pipeline (setup → prepare → train → evaluate → ollama)
    help                     Show this help message

EXAMPLES:
    $0 setup
    $0 prepare-data
    $0 train
    $0 evaluate models/final_model
    $0 create-ollama models/final_model
    $0 full-pipeline

DIRECTORIES:
    Data: $DATA_DIR
    Models: $MODELS_DIR
    Logs: $LOGS_DIR

For more information, see the README.md file.
EOF
}

# Main script logic
main() {
    local command="$1"
    shift

    case "$command" in
        setup)
            check_prerequisites
            install_dependencies
            print_success "Training environment setup complete."
            ;;
        prepare-data)
            check_prerequisites
            prepare_data
            ;;
        train)
            check_prerequisites
            train_model
            ;;
        evaluate)
            check_prerequisites
            evaluate_model "$@"
            ;;
        create-ollama)
            check_prerequisites
            create_ollama_model "$@"
            ;;
        full-pipeline)
            full_training_pipeline
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"