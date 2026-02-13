.PHONY: help postgres createdb dropdb migrateup migratedown sqlc test test-unit test-integration test-coverage clean run lint docker-up docker-down docker-rebuild user-service-build user-service-run api-gateway-build api-gateway-run logs logs-user logs-gateway

help:
	@echo "Available commands:"
	@echo "  make postgres            - Start PostgreSQL container"
	@echo "  make createdb            - Create database"
	@echo "  make dropdb              - Drop database"
	@echo "  make migrateup           - Run database migrations"
	@echo "  make migratedown         - Rollback database migrations"
	@echo "  make sqlc                - Generate Go code from SQL"
	@echo "  make test                - Run all tests"
	@echo "  make test-unit           - Run unit tests only"
	@echo "  make test-integration    - Run integration tests (needs Docker)"
	@echo "  make test-coverage       - Run tests with coverage report"
	@echo "  make lint                - Run linters"
	@echo "  make run                 - Run the monolith application"
	@echo "  make clean               - Stop containers and clean up"
	@echo ""
	@echo "Microservices commands:"
	@echo "  make docker-up           - Start all microservices with Docker Compose"
	@echo "  make docker-down         - Stop all microservices"
	@echo "  make docker-rebuild      - Rebuild and restart all services"
	@echo "  make user-service-build  - Build USER service"
	@echo "  make user-service-run    - Run USER service locally"
	@echo "  make api-gateway-build   - Build API Gateway service"
	@echo "  make api-gateway-run     - Run API Gateway locally"
	@echo "  make logs                - View logs from all services"
	@echo "  make logs-user           - View logs from USER service"
	@echo "  make logs-gateway        - View logs from API Gateway"

# Start PostgreSQL with Docker Compose
postgres:
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

# Create the database (if needed manually)
createdb:
	docker exec -it user_api_db createdb --username=postgres --owner=postgres user_management

# Drop the database
dropdb:
	docker exec -it user_api_db dropdb --username=postgres user_management

# Run migrations up
migrateup:
	migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/user_management?sslmode=disable" -verbose up

# Run migrations down
migratedown:
	migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/user_management?sslmode=disable" -verbose down

# Generate Go code from SQL using sqlc
sqlc:
	sqlc generate

# ============================================================================
# Testing Commands
# ============================================================================

# Run all tests (unit + integration)
test:
	go test -v -cover ./...

# Run only unit tests (skip integration tests)
# This runs tests in internal/ which don't need a database
test-unit:
	go test -v -cover ./internal/...

# Run integration tests (requires Docker with PostgreSQL running)
# These tests use a real database
test-integration:
	go test -v ./tests/integration/...

# Run tests with HTML coverage report
# Opens coverage.html in your browser
test-coverage:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
ifeq ($(OS),Windows_NT)
	start coverage.html
else
	open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || echo "Open coverage.html manually"
endif

# Run linters using golangci-lint
lint:
	golangci-lint run ./...

# Run the application
run:
	go run cmd/api/main.go

# Clean up - stop containers
clean:
	docker-compose down
	docker volume rm user-management-api_postgres_data || true

# Install dependencies
deps:
	go mod download
	go mod tidy

# ============================================================================
# Microservices Commands
# ============================================================================

# Start all services with Docker Compose
docker-up:
	docker-compose up -d --build
	@echo "All services started. Check status with: docker-compose ps"

# Stop all services
docker-down:
	docker-compose down
	@echo "All services stopped."

# Build USER service
user-service-build:
	cd services/user-service && go build -o bin/user-service ./cmd/main.go

# Run USER service locally
user-service-run:
	cd services/user-service && go run ./cmd/main.go

# Build API Gateway service
api-gateway-build:
	cd services/api-gateway && go build -o bin/api-gateway ./cmd/main.go

# Run API Gateway locally
api-gateway-run:
	cd services/api-gateway && go run ./cmd/main.go

# Rebuild and restart all services
docker-rebuild:
	docker-compose down
	docker-compose up -d --build
	@echo "All services rebuilt and started."

# View logs from all services
logs:
	docker-compose logs -f

# View logs from USER service only
logs-user:
	docker-compose logs -f user-service

# View logs from API Gateway only
logs-gateway:
	docker-compose logs -f api-gateway