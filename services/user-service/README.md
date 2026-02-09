# USER Service

The USER Service is a microservice responsible for managing user data. It handles all CRUD operations for users, communicates via NATS RPC, and publishes domain events when users are created, updated, or deleted.

## Architecture

```
┌──────────────────────┐
│    USER Service      │
│                      │
│  ┌────────────────┐  │
│  │   RPC Server   │◄─┼── NATS RPC (user.*)
│  └───────┬────────┘  │
│          │           │
│  ┌───────▼────────┐  │
│  │  User Service  │  │
│  └───────┬────────┘  │
│          │           │
│  ┌───────▼────────┐  │
│  │   Repository   │──┼── PostgreSQL
│  └────────────────┘  │
│          │           │
│  ┌───────▼────────┐  │
│  │Event Publisher │──┼── NATS PubSub (user.events.*)
│  └────────────────┘  │
└──────────────────────┘
```

## NATS Communication

### RPC Subjects (Request/Response)

| Subject | Description | Request | Response |
|---------|-------------|---------|----------|
| `user.create` | Create a new user | `CreateUserRPCRequest` | `UserResponse` |
| `user.get` | Get user by ID | `GetUserRPCRequest` | `UserResponse` |
| `user.list` | List all users | `ListUsersRPCRequest` | `ListUsersResponse` |
| `user.update` | Update a user | `UpdateUserRPCRequest` | `UserResponse` |
| `user.delete` | Delete a user | `DeleteUserRPCRequest` | `DeleteUserResponse` |

### PubSub Topics (Events)

| Topic | Description | Payload |
|-------|-------------|---------|
| `user.events.created` | Published when user is created | `Event` with `UserCreatedEventData` |
| `user.events.updated` | Published when user is updated | `Event` with `UserUpdatedEventData` |
| `user.events.deleted` | Published when user is deleted | `Event` with `UserDeletedEventData` |

### Event Payload Structure

```json
{
  "event_id": "uuid",
  "event_type": "user.created|user.updated|user.deleted",
  "timestamp": "2024-01-01T12:00:00Z",
  "actor_id": "uuid of user who made the change",
  "user_id": "uuid of affected user",
  "data": {
    "user": {
      "userId": "uuid",
      "firstName": "string",
      "lastName": "string",
      "email": "string",
      "phone": "string (optional)",
      "age": "number (optional)",
      "status": "Active|Inactive",
      "createdAt": "timestamp",
      "updatedAt": "timestamp"
    }
  }
}
```

## Configuration

The service is configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | Database name | `user_management` |
| `NATS_URL` | NATS server URL | `nats://localhost:4222` |
| `LOG_LEVEL` | Logging level | `info` |

## Project Structure

```
user-service/
├── cmd/
│   └── main.go              # Application entry point
├── db/
│   ├── migrations/          # Database migrations
│   ├── queries/             # SQL queries for sqlc
│   └── sqlc/                # Generated Go code
├── internal/
│   ├── config/              # Configuration loading
│   ├── models/              # Domain models and DTOs
│   ├── nats/
│   │   ├── publisher/       # Event publisher
│   │   └── rpc/             # RPC server and message types
│   ├── service/             # Business logic
│   ├── utils/               # Utility functions
│   └── validator/           # Input validation
├── Dockerfile               # Docker build file
├── Makefile                 # Build commands
├── go.mod                   # Go module definition
└── sqlc.yml                 # sqlc configuration
```

## Running Locally

### Prerequisites

- Go 1.25+
- PostgreSQL
- NATS Server

### Steps

1. Start PostgreSQL and NATS:
   ```bash
   docker-compose up -d postgres nats
   ```

2. Run migrations:
   ```bash
   make migrateup
   ```

3. Run the service:
   ```bash
   make run
   ```

## Running with Docker

```bash
# From the root directory
docker-compose up -d user-service
```

## Development

### Generate sqlc code

```bash
make sqlc
```

### Run tests

```bash
make test
```

### Run linter

```bash
make lint
```

## API Reference

### RPC Request/Response Examples

#### Create User

**Request:**
```json
{
  "actor_id": "550e8400-e29b-41d4-a716-446655440000",
  "data": {
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "age": 30,
    "status": "Active"
  }
}
```

**Response (Success):**
```json
{
  "success": true,
  "data": {
    "userId": "123e4567-e89b-12d3-a456-426614174000",
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "age": 30,
    "status": "Active",
    "createdAt": "2024-01-01T12:00:00Z",
    "updatedAt": "2024-01-01T12:00:00Z"
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": {
    "code": 409,
    "message": "Email already exists"
  }
}
```

#### Get User

**Request:**
```json
{
  "actor_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

#### Update User

**Request:**
```json
{
  "actor_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "data": {
    "firstName": "Jane",
    "status": "Inactive"
  }
}
```

#### Delete User

**Request:**
```json
{
  "actor_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```
