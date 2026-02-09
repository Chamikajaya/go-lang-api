# API Gateway Service

The API Gateway is the entry point for all client requests. It exposes REST endpoints and forwards requests to the USER Service via NATS RPC.

## Architecture

```
┌──────────────────────────────┐
│       API Gateway            │
│                              │
│  ┌────────────────────────┐  │
│  │     REST Handlers      │◄─┼── HTTP Requests
│  │  (Chi Router)          │  │
│  └───────────┬────────────┘  │
│              │               │
│  ┌───────────▼────────────┐  │
│  │     NATS RPC Client    │──┼── NATS RPC (user.*)
│  └────────────────────────┘  │
│                              │
│  ┌────────────────────────┐  │
│  │   Swagger/OpenAPI      │──┼── /swagger/*
│  └────────────────────────┘  │
└──────────────────────────────┘
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/users` | Create a new user |
| GET | `/users` | List all users |
| GET | `/users/{id}` | Get user by ID |
| PATCH | `/users/{id}` | Update a user |
| DELETE | `/users/{id}` | Delete a user |
| GET | `/health` | Health check |
| GET | `/swagger/*` | Swagger documentation |

## Configuration

The service is configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `NATS_URL` | NATS server URL | `nats://localhost:4222` |
| `RPC_TIMEOUT_SEC` | RPC call timeout in seconds | `5` |
| `LOG_LEVEL` | Logging level | `info` |

## Headers

| Header | Description | Required |
|--------|-------------|----------|
| `X-Actor-ID` | UUID of the user making the request | No (defaults to zeros) |
| `Content-Type` | Must be `application/json` for POST/PATCH | Yes |

## Project Structure

```
api-gateway/
├── cmd/
│   └── main.go              # Application entry point
├── docs/                    # Swagger documentation
├── internal/
│   ├── config/              # Configuration loading
│   ├── handlers/            # HTTP request handlers
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Request/response models
│   └── nats/
│       └── client/          # NATS RPC client
├── Dockerfile               # Docker build file
├── Makefile                 # Build commands
└── go.mod                   # Go module definition
```

## Running Locally

### Prerequisites

- Go 1.25+
- NATS Server
- USER Service running

### Steps

1. Start NATS and USER Service:
   ```bash
   docker-compose up -d nats user-service
   ```

2. Run the API Gateway:
   ```bash
   make run
   ```

3. Access Swagger documentation:
   ```
   http://localhost:8080/swagger/index.html
   ```

## Running with Docker

```bash
# From the root directory
docker-compose up -d api-gateway
```

## Testing with Postman

### Create User
```
POST http://localhost:8080/users
Content-Type: application/json

{
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "age": 30,
    "status": "Active"
}
```

### Get All Users
```
GET http://localhost:8080/users
```

### Get User by ID
```
GET http://localhost:8080/users/{userId}
```

### Update User
```
PATCH http://localhost:8080/users/{userId}
Content-Type: application/json

{
    "firstName": "Jane",
    "status": "Inactive"
}
```

### Delete User
```
DELETE http://localhost:8080/users/{userId}
```

## Development

### Generate Swagger docs

```bash
make swagger
```

### Run tests

```bash
make test
```

### Run linter

```bash
make lint
```

## Error Responses

All errors follow this format:

```json
{
    "error": "Bad Request",
    "message": "Invalid request body",
    "details": {
        "firstName": "firstName is required"
    }
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (validation error) |
| 404 | Not Found |
| 409 | Conflict (e.g., duplicate email) |
| 500 | Internal Server Error |
| 503 | Service Unavailable (USER service down) |
| 504 | Gateway Timeout (RPC timeout) |
