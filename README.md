# Astrology AI Backend

A clean Go backend architecture for an astrology AI application that provides birth chart calculations, horoscopes, and compatibility analysis.

## Architecture

This project follows Clean Architecture principles with the following structure:

```
cmd/
├── main.go                 # Application entry point
├── routes.go               # Route configuration and setup

internal/
├── config/                 # Configuration management
├── constants/              # Application constants
├── database/               # Database connection and migrations
├── handlers/               # HTTP request handlers
├── interfaces/             # Interface definitions for dependency injection
├── middleware/             # HTTP middleware
├── mocks/                  # Mock implementations for testing
├── models/                 # Data models
├── repositories/           # Data access layer
└── services/               # Business logic layer

pkg/                        # Public libraries and utilities
```

## Features

- **User Management**: Registration, authentication, and profile management
- **Birth Chart Calculation**: Generate astrological birth charts
- **Daily Horoscopes**: AI-generated daily horoscopes for zodiac signs (with Ollama integration)
- **Compatibility Analysis**: Check compatibility between birth charts
- **Astrological Remedies**: Personalized remedies and recommendations based on birth charts
- **JWT Authentication**: Secure API endpoints with JWT tokens
- **PostgreSQL Database**: Robust data persistence with automatic table creation
- **Comprehensive Testing**: Unit tests with mocks, integration tests, and end-to-end tests
- **Clean Architecture**: Interface-based design with dependency injection
- **AI Model Training**: Complete pipeline for fine-tuning custom astrology AI models

## AI Training Pipeline

This project includes a comprehensive AI training infrastructure for fine-tuning language models specifically for astrology content generation.

### Training Features

- **Data Preparation**: Automated data cleaning, formatting, and preprocessing
- **LoRA Fine-tuning**: Efficient parameter-efficient fine-tuning using LoRA adapters
- **Model Evaluation**: Comprehensive metrics including perplexity, BLEU, ROUGE, and astrology-specific accuracy
- **Ollama Integration**: Automatic conversion of trained models to Ollama format for easy deployment
- **Distributed Training**: Support for multi-GPU training with gradient accumulation

### Training Prerequisites

- Python 3.8+
- PyTorch 2.0+
- Transformers library
- PEFT (Parameter-Efficient Fine-Tuning)
- CUDA-compatible GPU (recommended for training)

### Quick Start Training

1. Setup training environment:
```bash
cd scripts/train
./train.sh setup
```

2. Prepare training data:
```bash
./train.sh prepare-data
```

3. Train the model:
```bash
./train.sh train
```

4. Evaluate the trained model:
```bash
./train.sh evaluate
```

5. Create Ollama model:
```bash
./train.sh create-ollama
```

Or run the complete pipeline:
```bash
./train.sh full-pipeline
```

### Training Scripts

- `prepare_data.py`: Data preprocessing and formatting
- `train_model.py`: LoRA fine-tuning with PyTorch/Transformers
- `evaluate_model.py`: Comprehensive model evaluation with multiple metrics
- `create_ollama_model.py`: Convert trained model to Ollama format

### Training Configuration

Training parameters are configured in `scripts/train/config/`:
- `data_config.yaml`: Data preprocessing settings
- `train_config.yaml`: Model training hyperparameters
- `eval_config.yaml`: Evaluation metrics and settings

### Sample Data

Sample training data is provided in `data/raw/astrology_training_data.json` with examples of:
- Daily horoscopes for all zodiac signs
- Compatibility analysis between signs
- Astrological remedies and recommendations

### Custom Training Data

To use your own training data:
1. Add JSON files to `data/raw/` directory
2. Format as instruction-response pairs
3. Run `prepare_data.py` to process the data

### Model Deployment

After training, the model can be deployed with Ollama:
```bash
# Install the trained model
cd ollama_models
./install.sh

# Use the model
ollama run astrology-ai
```

### Training Metrics

The evaluation provides:
- **Perplexity**: Language model quality metric
- **BLEU/ROUGE**: Text generation quality metrics
- **Astrology Accuracy**: Custom metric for astrology-specific content quality
- **Response Length**: Average and distribution analysis

## Prerequisites

- Go 1.23 or later
- PostgreSQL database
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/vikhyat-sharma/astrology-ai.git
cd astrology-ai
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
Create a `.env` file in the root directory:
```env
DATABASE_URL=postgres://postgres:password@localhost:5432/astrology_ai?sslmode=disable
JWT_SECRET=your-secret-key
PORT=8080
ENVIRONMENT=development
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama2
```

4. Set up the database:
The application will automatically create the database and tables on startup. Ensure PostgreSQL is running and accessible.

For local development, you can use Docker:
```bash
# Start PostgreSQL with Docker
docker run --name postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=astrology_ai -p 5432:5432 -d postgres:15-alpine

# Or use docker-compose
docker-compose up -d db
```

5. Run the application:
```bash
go run cmd/main.go cmd/routes.go
```

Or use the Makefile:
```bash
make run
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user

### User (Protected)
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update user profile

### Astrology (Protected)
- `POST /api/v1/astrology/birth-chart` - Create birth chart
- `GET /api/v1/astrology/birth-chart/:id` - Get birth chart
- `GET /api/v1/astrology/horoscope/daily?sign=Leo` - Get AI-generated daily horoscope
- `POST /api/v1/astrology/compatibility` - Check compatibility
- `GET /api/v1/astrology/remedies/:id` - Get personalized astrological remedies

### Health Check
- `GET /health` - Health check endpoint

## API Usage Examples

### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Create Birth Chart (requires Bearer token)
```bash
curl -X POST http://localhost:8080/api/v1/astrology/birth-chart \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "birth_date": "1990-01-15",
    "birth_time": "14:30",
    "birth_place": "New York, NY",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "timezone": "America/New_York"
  }'
```

### Get Daily Horoscope (requires Bearer token)
```bash
curl -X GET "http://localhost:8080/api/v1/astrology/horoscope/daily?sign=Leo" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get Astrological Remedies (requires Bearer token)
```bash
curl -X GET "http://localhost:8080/api/v1/astrology/remedies/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Development

### Makefile Commands

This project includes a comprehensive Makefile for common development tasks:

```bash
# Show all available commands
make help

# Testing
make test              # Run all tests
make test-unit         # Run unit tests only
make test-integration  # Run integration tests only
make test-e2e          # Run end-to-end tests only
make test-coverage     # Run tests with coverage report

# Development
make run               # Run the application
make build             # Build the application
make clean             # Clean build artifacts
make fmt               # Format Go code
make vet               # Run go vet
make lint              # Run golangci-lint

# Docker
make docker-build      # Build Docker image
make docker-run        # Run Docker container
make docker-compose-up # Start services with docker-compose

# Dependencies
make deps              # Download dependencies
make deps-update       # Update dependencies
make tidy              # Tidy and verify module dependencies

# Setup
make setup             # Setup development environment
make install-tools     # Install development tools

# Quality checks
make check             # Run format, vet, lint and tests
make ci                # Run CI pipeline locally
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test types
go test ./internal/services/...     # Unit tests with mocks
go test ./internal/services/*integration_test.go  # Integration tests
go test ./internal/e2e/...          # End-to-end tests
```

### Testing Architecture

The project includes a comprehensive testing suite with three levels:

- **Unit Tests**: Test individual functions and methods using mock dependencies
- **Integration Tests**: Test component interactions with SQLite in-memory database
- **End-to-End Tests**: Test complete API workflows with real HTTP requests

Mock implementations are provided for external dependencies like HTTP clients and repositories, enabling fast and reliable unit testing.

### Building
```bash
go build -o bin/astrology-ai cmd/main.go cmd/routes.go
```

Or use the Makefile:
```bash
make build
```

### Code Organization

- **Handlers**: Handle HTTP requests and responses, input validation
- **Services**: Contain business logic, orchestrate data operations with dependency injection
- **Repositories**: Handle data persistence and retrieval
- **Models**: Define data structures and database schemas
- **Interfaces**: Define contracts for dependency injection and mocking
- **Mocks**: Provide mock implementations for comprehensive testing
- **Middleware**: Cross-cutting concerns like authentication, logging, CORS
- **Config**: Centralized configuration management
- **Database**: Database connection and automatic schema management

## Database Schema

The application uses the following main entities (tables are created automatically on startup):

- **Users**: User accounts with authentication info
- **BirthCharts**: Astrological birth chart data with planetary positions
- **Horoscopes**: Daily/weekly horoscopes by zodiac sign with AI-generated content

Database migrations are handled automatically with fallback to manual SQL table creation for reliability.

## Security

- Passwords are hashed using bcrypt
- JWT tokens for authentication
- Input validation on all endpoints
- CORS protection
- SQL injection prevention with GORM

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.