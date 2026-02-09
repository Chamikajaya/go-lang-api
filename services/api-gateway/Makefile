# API Gateway Makefile

.PHONY: help build run test lint swagger clean docker-build deps

help:
	@echo "Available commands:"
	@echo "  make build        - Build the service binary"
	@echo "  make run          - Run the service locally"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linters"
	@echo "  make swagger      - Generate Swagger documentation"
	@echo "  make docker-build - Build Docker image"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"

# Build the service
build:
	go build -o bin/api-gateway ./cmd/main.go

# Run the service locally
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v -cover ./...

# Run linters
lint:
	golangci-lint run ./...

# Generate Swagger documentation
swagger:
	swag init -g cmd/main.go -o docs

# Build Docker image
docker-build:
	docker build -t api-gateway:latest .

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy
