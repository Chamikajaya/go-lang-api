.PHONY: help postgres createdb dropdb migrateup migratedown sqlc test clean run

help:
	@echo "Available commands:"
	@echo "  make postgres     - Start PostgreSQL container"
	@echo "  make createdb     - Create database"
	@echo "  make dropdb       - Drop database"
	@echo "  make migrateup    - Run database migrations"
	@echo "  make migratedown  - Rollback database migrations"
	@echo "  make sqlc         - Generate Go code from SQL"
	@echo "  make test         - Run tests"
	@echo "  make run          - Run the application"
	@echo "  make clean        - Stop containers and clean up"

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

# Run tests
test:
	go test -v -cover ./...

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