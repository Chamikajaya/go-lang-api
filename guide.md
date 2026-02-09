# MASTER PROMPT: Microservices Refactoring with API Gateway, NATS, and WebSockets

## PROJECT CONTEXT

You are refactoring an existing monolithic Go REST API for user management into a microservices architecture. The current implementation has CRUD operations for users using Chi Router, PostgreSQL with sqlc, and Go Playground Validator.

## CURRENT STATE

**Existing Technology Stack:**
- Language: Go
- Router: Chi Router
- Database: PostgreSQL (Docker)
- ORM: sqlc
- Validation: Go Playground Validator
- Testing: Go testing library
- Docs: OpenAPI
- Linting: golangci-lint

**Existing User Entity:**
```
- userId: UUID (Primary Key, Auto-generated)
- firstName: String (Required, Min 2, Max 50)
- lastName: String (Required, Min 2, Max 50)
- email: String (Required, Valid Email)
- phone: String (Optional, Valid Phone)
- age: Integer (Optional, Positive)
- status: Enum (Active, Inactive) (Optional, Default: Active)
```

**Existing Endpoints:**
- POST /users - Create user
- GET /users - Get all users
- GET /users/{id} - Get user by ID
- PATCH /users/{id} - Update user
- DELETE /users/{id} - Delete user

## TARGET ARCHITECTURE

**New Microservices Architecture:**

```
┌──────┐ REST/WS   ┌─────────────────────┐ NATS RPC    ┌──────────────────┐
│Client│◄─────────►│  API Gateway SVC    │◄───────────►│   USER SVC       │
└──────┘           │                     │             │                  │
                   │ • REST endpoints    │             │ • Business logic │
                   │ • WebSocket server  │             │ • DB operations  │
                   │ • NATS RPC client   │             │ • NATS RPC server│
                   │ • NATS PubSub sub   │             │ • NATS PubSub pub│
                   └──────────┬──────────┘             └────────┬─────────┘
                              │                                 │
                              │                                 ▼
                              │                          ┌─────────────┐
                              │                          │ PostgreSQL  │
                              │                          │  (USER DB)  │
                              │                          └─────────────┘
                              ▼
                      ┌──────────────────┐
                      │   NATS Broker    │
                      │                  │
                      │ • RPC channels   │
                      │ • PubSub topics  │
                      └──────────────────┘
```

## ARCHITECTURAL REQUIREMENTS

### 1. API Gateway Service
**Responsibilities:**
- Accept REST API requests from clients
- Manage WebSocket connections
- Forward requests to USER Service via NATS RPC
- Subscribe to NATS events (user.created, user.updated, user.deleted)
- Broadcast WebSocket messages to connected clients
- Handle request/response transformation

**Communication Patterns:**
- **REST → NATS RPC**: Convert REST requests to NATS RPC calls
- **NATS PubSub → WebSocket**: Listen to events and broadcast to WS clients

### 2. USER Service
**Responsibilities:**
- Handle all user business logic
- Perform database CRUD operations
- Validate input data
- Respond to NATS RPC requests
- Publish events to NATS PubSub after successful operations
- Maintain data integrity

**Communication Patterns:**
- **NATS RPC Server**: Handle requests from API Gateway
- **NATS PubSub Publisher**: Publish events after DB operations

### 3. NATS Communication Design

**RPC Subjects (Request/Response):**
- `user.create` - Create user RPC
- `user.get` - Get user by ID RPC
- `user.list` - Get all users RPC
- `user.update` - Update user RPC
- `user.delete` - Delete user RPC

**PubSub Topics (Events):**
- `user.events.created` - Published when user is created
- `user.events.updated` - Published when user is updated
- `user.events.deleted` - Published when user is deleted

**Event Payload Structure:**
```json
{
  "event_id": "uuid",
  "event_type": "user.created|user.updated|user.deleted",
  "timestamp": "2024-01-01T12:00:00Z",
  "actor_id": "uuid of user who made the change",
  "user_id": "uuid of affected user",
  "data": {
    // User data or changes
  }
}
```

### 4. WebSocket Requirements

**WebSocket Endpoint:**
- `WS /ws` - WebSocket connection endpoint

**Connection Management:**
- Clients connect with userId as query parameter: `/ws?userId={uuid}`
- Maintain map of userId → WebSocket connection
- Handle reconnections and heartbeat/ping-pong
- Clean up disconnected clients

**Message Broadcasting Rules:**
- When user is UPDATED: Send WS message to ALL users EXCEPT the user who made the update
- When user is CREATED: Send WS message to ALL connected users
- When user is DELETED: Send WS message to ALL connected users

**WebSocket Message Format:**
```json
{
  "type": "user.created|user.updated|user.deleted",
  "timestamp": "2024-01-01T12:00:00Z",
  "user_id": "uuid",
  "data": {
    // User data or update details
  }
}
```

## TECHNICAL SPECIFICATIONS

### Error Handling
- Maintain proper HTTP status codes (200, 201, 400, 404, 500)
- NATS RPC errors should be properly propagated
- WebSocket errors should be logged but not crash server
- Implement timeout for RPC calls (5 seconds)

### Configuration
- Use environment variables for configuration
- Support the following config:
  - Database connection string
  - NATS server URL
  - API Gateway port
  - USER Service port
  - Log level



### Code Quality
- Follow Go best practices and idioms
- Use proper error handling (no panic in production code)
- Pass golangci-lint with no errors
- Use context for cancellation and timeouts

### Docker Setup
- Dockerfile for API Gateway Service
- Dockerfile for USER Service
- Docker Compose file including:
  - PostgreSQL
  - NATS
  - API Gateway Service
  - USER Service



## IMPLEMENTATION PHASES

The implementation will be divided into the following phases:

### **Phase 1: Project Setup and Restructuring**
- Create new directory structure
- Set up Go modules for each service
- Configure Docker Compose with PostgreSQL and NATS
- Migrate existing code to new structure

### **Phase 2: USER Service - Core Implementation**
- Implement repository layer with sqlc
- Implement service layer with business logic
- Set up NATS RPC server
- Implement RPC handlers for all CRUD operations
- Add validation logic

### **Phase 3: USER Service - Event Publishing**
- Implement NATS PubSub publisher
- Publish events after successful CRUD operations
- Add event models and serialization
- Implement error handling for event publishing

### **Phase 4: API Gateway - REST & NATS RPC Client**
- Implement REST handlers using Chi Router
- Implement NATS RPC client
- Forward REST requests to USER Service via RPC
- Handle responses and errors
- Maintain OpenAPI documentation

### **Phase 5: API Gateway - WebSocket Implementation**
- Implement WebSocket server
- Create connection manager
- Handle client connections and disconnections
- Implement heartbeat/ping-pong mechanism

### **Phase 6: API Gateway - Event Subscription & Broadcasting**
- Subscribe to NATS PubSub events
- Implement broadcasting logic
- Filter messages (exclude actor from updates)
- Handle WebSocket errors gracefully

### **Phase 7: Testing & Documentation**
- Update OpenAPI documentation
- Create README with setup instructions

### **Phase 8: Docker & Deployment**
- Create Dockerfiles for both services
- Update Docker Compose
- Add health checks
- Create Makefile for common tasks
- Test full deployment

## SUCCESS CRITERIA

✅ API Gateway successfully forwards REST requests to USER Service via NATS RPC
✅ USER Service performs CRUD operations and publishes events
✅ WebSocket connections are stable and managed properly
✅ Events are broadcast to appropriate clients (excluding actor for updates)
✅ Docker Compose brings up all services successfully
✅ OpenAPI documentation is complete and accurate

## CONSTRAINTS AND GUIDELINES

1. **Preserve existing functionality**: All existing REST endpoints must work exactly as before
2. **Use existing validation**: Keep Go Playground Validator for input validation
3. **Database schema**: Do not modify the existing User table schema
4. **Response codes**: Maintain HTTP status codes as specified in original requirements
5. **Error messages**: Provide clear, actionable error messages
6. **Configuration**: Use 12-factor app principles
7. **Security**: Validate all inputs, sanitize outputs
8. **Performance**: RPC calls should have timeout, WebSocket should handle 1000+ concurrent connections

## EXPECTED OUTPUT

For each phase, provide:
1. Complete, production-ready code
2. Clear comments explaining complex logic
3. Error handling for all failure scenarios
6. Configuration examples
7.  explanation of implementation decisions

## NOTES

- This is a refactoring project, so reuse existing code where possible
- Focus on clean separation of concerns
- Make services independently deployable
- WebSocket is for notifications only, not for bidirectional communication
- NATS handles both synchronous (RPC) and asynchronous (PubSub) communication

---

