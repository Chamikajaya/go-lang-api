# User Service Makefile

.PHONY: help build run test lint sqlc clean docker-build deps

help:
	@echo "Available commands:"
	@echo "  make build        - Build the service binary"
	@echo "  make run          - Run the service locally"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linters"
	@echo "  make sqlc         - Generate Go code from SQL"
	@echo "  make docker-build - Build Docker image"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"

# Build the service
build:
	go build -o bin/user-service ./cmd/main.go

# Run the service locally
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v -cover ./...

# Run linters
lint:
	golangci-lint run ./...

# Generate sqlc code
sqlc:
	sqlc generate

# Build Docker image
docker-build:
	docker build -t user-service:latest .

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy
