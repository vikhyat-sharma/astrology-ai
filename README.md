# Astrology AI Backend

A clean Go backend architecture for an astrology AI application that provides birth chart calculations, horoscopes, and compatibility analysis.

## Architecture

This project follows Clean Architecture principles with the following structure:

```
cmd/
├── main.go                 # Application entry point

internal/
├── config/                 # Configuration management
├── database/               # Database connection and migrations
├── handlers/               # HTTP request handlers
├── middleware/             # HTTP middleware
├── models/                 # Data models
├── repositories/           # Data access layer
└── services/               # Business logic layer

pkg/                        # Public libraries and utilities
```

## Features

- **User Management**: Registration, authentication, and profile management
- **Birth Chart Calculation**: Generate astrological birth charts
- **Daily Horoscopes**: AI-generated daily horoscopes for zodiac signs
- **Compatibility Analysis**: Check compatibility between birth charts
- **JWT Authentication**: Secure API endpoints with JWT tokens
- **PostgreSQL Database**: Robust data persistence with GORM

## Prerequisites

- Go 1.21 or later
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
DATABASE_URL=postgres://user:password@localhost/astrology_ai?sslmode=disable
JWT_SECRET=your-secret-key
PORT=8080
ENVIRONMENT=development
```

4. Set up the database:
```bash
createdb astrology_ai
```

5. Run the application:
```bash
go run cmd/main.go
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
- `GET /api/v1/astrology/horoscope/daily?sign=Leo` - Get daily horoscope
- `POST /api/v1/astrology/compatibility` - Check compatibility

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
go test ./...
```

### Building
```bash
go build -o bin/astrology-ai cmd/main.go
```

### Code Organization

- **Handlers**: Handle HTTP requests and responses, input validation
- **Services**: Contain business logic, orchestrate data operations
- **Repositories**: Handle data persistence and retrieval
- **Models**: Define data structures and database schemas
- **Middleware**: Cross-cutting concerns like authentication, logging, CORS
- **Config**: Centralized configuration management

## Database Schema

The application uses the following main entities:

- **Users**: User accounts with authentication info
- **BirthCharts**: Astrological birth chart data
- **Horoscopes**: Daily/weekly horoscopes by zodiac sign

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