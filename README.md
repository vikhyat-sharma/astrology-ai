# 🔮 Astrology AI Backend

A production-grade Go backend for an AI-powered astrology platform—supporting birth chart calculations, horoscopes, compatibility analysis, and custom-trained language models.

Designed with clean architecture, scalability, and AI extensibility in mind.

---

## 🚀 Overview

This project provides a modular REST API for astrology-based services powered by AI. It combines:

* Structured backend engineering (Go + Clean Architecture)
* AI-driven content generation (Ollama + fine-tuned models)
* A complete ML training pipeline for domain-specific models

---

## ✨ Core Features

### 🧑‍💼 User & Auth

* User registration & login
* JWT-based authentication
* Profile management

### 🔭 Astrology Engine

* Birth chart generation (planetary positions, metadata)
* Daily AI-generated horoscopes
* Compatibility analysis between individuals
* Personalized astrological remedies

### 🤖 AI Integration

* Ollama-powered inference
* Custom fine-tuned astrology models
* Domain-specific evaluation metrics

### 🧪 Engineering Quality

* Clean Architecture with dependency injection
* Unit, integration, and E2E testing
* Mock-based testing support
* Docker-ready setup
* Makefile-driven workflows

---

## 🏗 Project Structure

```bash
cmd/
├── main.go        # Entry point
├── routes.go      # Route definitions

internal/
├── config/        # Configuration
├── constants/     # App-wide constants
├── database/      # DB setup & migrations
├── handlers/      # HTTP handlers
├── interfaces/    # Contracts for DI
├── middleware/    # HTTP middleware
├── mocks/         # Test mocks
├── models/        # Data models
├── repositories/  # Data layer
└── services/      # Business logic

pkg/               # Shared utilities
```

### Layer Responsibilities

* **Handlers** → Request parsing & response formatting
* **Services** → Core business logic
* **Repositories** → Database access
* **Interfaces** → Enable loose coupling & testing
* **Middleware** → Cross-cutting concerns (auth, logging)

---

## ⚙️ Prerequisites

* Go 1.23+
* PostgreSQL
* Git
* (Optional) Docker
* (Optional for AI training) Python 3.8+, CUDA GPU

---

## 🛠 Installation

```bash
# Clone repo
git clone https://github.com/vikhyat-sharma/astrology-ai.git
cd astrology-ai

# Install dependencies
go mod download
```

### 🔑 Environment Variables

Create a `.env` file:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/astrology_ai?sslmode=disable
JWT_SECRET=your-secret-key
PORT=8080
ENVIRONMENT=development
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama2
```

---

## 🗄 Database Setup

Run PostgreSQL locally or via Docker:

```bash
docker run --name postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=astrology_ai \
  -p 5432:5432 -d postgres:15-alpine
```

> Tables are created automatically on startup.

---

## ▶️ Running the App

```bash
go run cmd/main.go cmd/routes.go
```

Or:

```bash
make run
```

App runs at: `http://localhost:8080`

---

## 🐳 Docker

```bash
make docker-compose-up
```

---

## 📡 API Endpoints

### Auth

* `POST /api/v1/auth/register`
* `POST /api/v1/auth/login`

### User (Protected)

* `GET /api/v1/user/profile`
* `PUT /api/v1/user/profile`

### Astrology (Protected)

* `POST /api/v1/astrology/birth-chart`
* `GET /api/v1/astrology/birth-chart/:id`
* `GET /api/v1/astrology/horoscope/daily?sign=Leo`
* `POST /api/v1/astrology/compatibility`
* `GET /api/v1/astrology/remedies/:id`

### Health

* `GET /health`

---

## 🧪 Testing

```bash
make test
make test-coverage
```

### Test Types

* **Unit Tests** → Fast, mock-based
* **Integration Tests** → DB-backed (SQLite in-memory)
* **E2E Tests** → Full API validation

---

## 🤖 AI Training Pipeline

A complete pipeline for fine-tuning astrology-specific LLMs.

### Features

* Data preprocessing & formatting
* LoRA-based fine-tuning (efficient & scalable)
* Multi-metric evaluation:

  * Perplexity
  * BLEU / ROUGE
  * Custom astrology accuracy
* Ollama model packaging
* Distributed training support

---

### ⚡ Quick Start

```bash
cd scripts/train

./train.sh setup
./train.sh prepare-data
./train.sh train
./train.sh evaluate
./train.sh create-ollama
```

Or run everything:

```bash
./train.sh full-pipeline
```

---

### 📁 Training Components

* `prepare_data.py` → Cleans & formats data
* `train_model.py` → LoRA fine-tuning
* `evaluate_model.py` → Metrics & evaluation
* `create_ollama_model.py` → Deployment packaging

---

### 📊 Training Data

Located in:

```
data/raw/astrology_training_data.json
```

Supports:

* Horoscope generation
* Compatibility analysis
* Remedies & recommendations

---

### 🧩 Custom Data

1. Add JSON files to `data/raw/`
2. Format as instruction-response pairs
3. Run preprocessing

---

## 🚀 Deployment with Ollama

```bash
cd ollama_models
./install.sh

ollama run astrology-ai
```

---

## 🔐 Security

* bcrypt password hashing
* JWT authentication
* Input validation
* SQL injection protection (ORM-based)
* CORS middleware

---

## 🧰 Development Workflow

```bash
make help        # List commands
make check       # Run lint + tests
make ci          # Simulate CI pipeline
```

---

## 🧱 Database Entities

* **Users**
* **BirthCharts**
* **Horoscopes**

> Auto-migrated on startup with fallback mechanisms.

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit changes
4. Add/modify tests
5. Open a Pull Request

---

## 📄 License

MIT License

---

## 💡 Future Improvements

* Real-time transit tracking
* Multi-language horoscope generation
* WebSocket-based live predictions
* GraphQL API layer
* Advanced ML personalization

---

## 🧭 Final Note

This project is more than a backend—it’s a foundation for building intelligent astrology platforms with modern AI capabilities. Extend it with your own models, datasets, or product ideas to unlock its full potential.
