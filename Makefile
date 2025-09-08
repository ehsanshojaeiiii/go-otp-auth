.PHONY: build run test clean docker-up docker-down docker-build swagger deps

# Variables
APP_NAME=golang-otp-service
DOCKER_IMAGE=$(APP_NAME)
MAIN_PATH=cmd/main.go

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) $(MAIN_PATH)

# Run the application locally
run:
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Generate swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g $(MAIN_PATH) -o ./docs

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

docker-down:
	@echo "Stopping services..."
	@docker-compose down

docker-logs:
	@echo "Showing logs..."
	@docker-compose logs -f

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development setup complete!"

# Run with Docker (full stack)
docker-run: docker-build docker-up

# Database operations
db-up:
	@echo "Starting database services..."
	@docker-compose up -d postgres redis

db-down:
	@echo "Stopping database services..."
	@docker-compose stop postgres redis

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  deps         - Install dependencies"
	@echo "  swagger      - Generate Swagger documentation"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start all services with Docker Compose"
	@echo "  docker-down  - Stop all services"
	@echo "  docker-logs  - Show Docker logs"
	@echo "  docker-run   - Build and run with Docker"
	@echo "  dev-setup    - Setup development environment"
	@echo "  db-up        - Start database services only"
	@echo "  db-down      - Stop database services"
	@echo "  help         - Show this help message"
