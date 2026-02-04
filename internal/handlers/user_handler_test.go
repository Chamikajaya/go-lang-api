// Package handlers_test contains unit tests for HTTP handlers.
//
// KEY CONCEPT: Testing HTTP Handlers
// ===================================
// To test HTTP handlers without starting a real server, Go provides:
// - httptest.NewRequest: Creates a fake HTTP request
// - httptest.NewRecorder: Records the HTTP response
//
// We can then check the response status code, body, headers, etc.
package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	database "user-management-api/db/sqlc"
	"user-management-api/internal/handlers"
	"user-management-api/internal/models"
	"user-management-api/internal/service"
	"user-management-api/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================================
// Mock Implementation (same as service_test, needed for handler tests)
// ============================================================================

type MockQuerier struct {
	CreateUserFunc     func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserByIDFunc    func(ctx context.Context, userID uuid.UUID) (database.User, error)
	ListUsersFunc      func(ctx context.Context) ([]database.User, error)
	UpdateUserFunc     func(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
	DeleteUserFunc     func(ctx context.Context, userID uuid.UUID) error
	EmailExistsFunc    func(ctx context.Context, email string) (bool, error)
	UserExistsFunc     func(ctx context.Context, userID uuid.UUID) (bool, error)
	GetUserByEmailFunc func(ctx context.Context, email string) (database.User, error)
	ListUsersByStatusFunc func(ctx context.Context, status string) ([]database.User, error)
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, arg)
	}
	return database.User{}, nil
}

func (m *MockQuerier) GetUserByID(ctx context.Context, userID uuid.UUID) (database.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return database.User{}, nil
}

func (m *MockQuerier) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return database.User{}, nil
}

func (m *MockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return []database.User{}, nil
}

func (m *MockQuerier) ListUsersByStatus(ctx context.Context, status string) ([]database.User, error) {
	if m.ListUsersByStatusFunc != nil {
		return m.ListUsersByStatusFunc(ctx, status)
	}
	return []database.User{}, nil
}

func (m *MockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, arg)
	}
	return database.User{}, nil
}

func (m *MockQuerier) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockQuerier) EmailExists(ctx context.Context, email string) (bool, error) {
	if m.EmailExistsFunc != nil {
		return m.EmailExistsFunc(ctx, email)
	}
	return false, nil
}

func (m *MockQuerier) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	if m.UserExistsFunc != nil {
		return m.UserExistsFunc(ctx, userID)
	}
	return false, nil
}

var _ database.Querier = (*MockQuerier)(nil)

// ============================================================================
// Helper Functions
// ============================================================================

// setupHandler creates a handler with mocked dependencies
func setupHandler(mock *MockQuerier) *handlers.UserHandler {
	queries := database.New(mock)
	userService := service.NewUserService(nil, queries)
	validatorInstance := validator.NewValidator()
	return handlers.NewUserHandler(userService, validatorInstance)
}

// createMockUser creates a mock database user
func createMockUser(id uuid.UUID, firstName, lastName, email string) database.User {
	return database.User{
		UserID:    id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     pgtype.Text{Valid: false},
		Age:       pgtype.Int4{Valid: false},
		Status:    "Active",
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

// ============================================================================
// CreateUser Handler Tests
// ============================================================================

func TestCreateUser_Success(t *testing.T) {
	mock := &MockQuerier{
		EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		CreateUserFunc: func(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
			return createMockUser(uuid.New(), arg.FirstName, arg.LastName, arg.Email), nil
		},
	}

	handler := setupHandler(mock)

	// Create request body
	// json.Marshal converts a Go struct to JSON bytes
	body, _ := json.Marshal(models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	})

	// Create a fake HTTP request
	// httptest.NewRequest creates a request without starting a server
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to capture the response
	// It implements http.ResponseWriter
	rr := httptest.NewRecorder()

	// Call the handler
	handler.CreateUser(rr, req)

	// Check status code
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	// Parse response body
	var response models.UserResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.FirstName != "John" {
		t.Errorf("Expected FirstName 'John', got '%s'", response.FirstName)
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	handler := setupHandler(&MockQuerier{})

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("not valid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	// Should return 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCreateUser_ValidationError(t *testing.T) {
	handler := setupHandler(&MockQuerier{})

	// Missing required fields
	body, _ := json.Marshal(models.CreateUserRequest{
		FirstName: "", // Required but empty
		LastName:  "", // Required but empty
		Email:     "", // Required but empty
	})

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	// Should return 400 Bad Request for validation errors
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// Check that response contains validation error details
	var response models.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error != "Validation Failed" {
		t.Errorf("Expected 'Validation Failed', got '%s'", response.Error)
	}
}

func TestCreateUser_EmailConflict(t *testing.T) {
	mock := &MockQuerier{
		EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Email already exists
		},
	}

	handler := setupHandler(mock)

	body, _ := json.Marshal(models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "existing@example.com",
	})

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	// Should return 409 Conflict
	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

// ============================================================================
// GetUser Handler Tests
// ============================================================================

func TestGetUser_Success(t *testing.T) {
	userID := uuid.New()
	mockUser := createMockUser(userID, "Jane", "Doe", "jane@example.com")

	mock := &MockQuerier{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
			return mockUser, nil
		},
	}

	handler := setupHandler(mock)

	// Create request with URL parameter
	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)

	// Chi router needs the URL param to be set in context
	// This is how chi passes route parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.UserResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.FirstName != "Jane" {
		t.Errorf("Expected 'Jane', got '%s'", response.FirstName)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	mock := &MockQuerier{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
			return database.User{}, pgx.ErrNoRows
		},
	}

	handler := setupHandler(mock)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestGetUser_InvalidUUID(t *testing.T) {
	handler := setupHandler(&MockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// ============================================================================
// ListUsers Handler Tests
// ============================================================================

func TestListUsers_Success(t *testing.T) {
	mockUsers := []database.User{
		createMockUser(uuid.New(), "John", "Doe", "john@example.com"),
		createMockUser(uuid.New(), "Jane", "Smith", "jane@example.com"),
	}

	mock := &MockQuerier{
		ListUsersFunc: func(ctx context.Context) ([]database.User, error) {
			return mockUsers, nil
		},
	}

	handler := setupHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.ListUsersResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Total != 2 {
		t.Errorf("Expected 2 users, got %d", response.Total)
	}
}

func TestListUsers_Empty(t *testing.T) {
	mock := &MockQuerier{
		ListUsersFunc: func(ctx context.Context) ([]database.User, error) {
			return []database.User{}, nil
		},
	}

	handler := setupHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.ListUsersResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Total != 0 {
		t.Errorf("Expected 0 users, got %d", response.Total)
	}
}

// ============================================================================
// DeleteUser Handler Tests
// ============================================================================

func TestDeleteUser_Success(t *testing.T) {
	userID := uuid.New()

	mock := &MockQuerier{
		UserExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
		DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	handler := setupHandler(mock)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.DeleteUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.SuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Message != "User deleted successfully" {
		t.Errorf("Expected success message, got '%s'", response.Message)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	mock := &MockQuerier{
		UserExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, nil // User doesn't exist
		},
	}

	handler := setupHandler(mock)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.DeleteUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}
