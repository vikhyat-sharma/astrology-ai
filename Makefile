# Makefile for Astrology AI Backend

.PHONY: help test test-unit test-integration test-e2e run build clean docker-build docker-run migrate lint fmt vet deps setup

# Default target
help: ## Show this help message
	@echo "Astrology AI Backend - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

# Testing commands
test: ## Run all tests
	go test ./...

test-unit: ## Run unit tests only
	go test ./internal/services -v

test-integration: ## Run integration tests only
	go test ./internal/services -run "Integration" -v

test-e2e: ## Run end-to-end tests only
	go test ./internal/e2e -v

test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Development commands
run: ## Run the application
	go run cmd/main.go

build: ## Build the application
	go build -o bin/astrology-ai cmd/main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# Docker commands
docker-build: ## Build Docker image
	docker build -t astrology-ai .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env astrology-ai

docker-compose-up: ## Start services with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	docker-compose down

# Database commands
migrate: ## Run database migrations (auto-migrate on startup)
	@echo "Database migrations run automatically on application startup"

db-reset: ## Reset database (for development)
	@echo "Resetting database..."
	# Add database reset logic here if needed

# Code quality commands
lint: ## Run golangci-lint (install with: make install-tools)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: make install-tools"; \
		echo "Skipping lint check..."; \
	fi

fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy and verify module dependencies
	go mod tidy
	go mod verify

# Dependency management
deps: ## Download dependencies
	go mod download

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

# Development setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment setup complete!"

# Utility commands
check: fmt vet lint test ## Run format, vet, lint and tests

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest

# CI/CD commands
ci: check ## Run CI pipeline locally
	@echo "CI pipeline completed successfully"

# Quick development workflow
dev: ## Start development server with hot reload (requires air)
	air

# Production build
prod-build: clean deps ## Build for production
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/astrology-ai cmd/main.go

# Help for specific targets
help-test: ## Show testing help
	@echo "Testing commands:"
	@echo "  make test              - Run all tests"
	@echo "  make test-unit         - Run unit tests only"
	@echo "  make test-integration  - Run integration tests only"
	@echo "  make test-e2e          - Run end-to-end tests only"
	@echo "  make test-coverage     - Run tests with coverage report"

help-docker: ## Show Docker help
	@echo "Docker commands:"
	@echo "  make docker-build      - Build Docker image"
	@echo "  make docker-run        - Run Docker container"
	@echo "  make docker-compose-up - Start services with docker-compose"
	@echo "  make docker-compose-down - Stop services with docker-compose"