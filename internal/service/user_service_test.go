// Package service_test contains unit tests for the UserService.
//
// KEY CONCEPT: Mocking
// ====================
// The UserService depends on the database (queries *database.Queries).
// In unit tests, we don't want to use a real database because:
// 1. It would be slow
// 2. Tests would fail if DB is not running
// 3. Tests might interfere with each other
//
// Instead, we create a "mock" - a fake implementation that we control.
// SQLC generated a Querier interface (in db/sqlc/querier.go) that we can mock!
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	database "user-management-api/db/sqlc"
	"user-management-api/internal/models"
	"user-management-api/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================================
// Mock Implementation
// ============================================================================

// MockQuerier is our fake database that implements the Querier interface.
// We control what it returns, so we can test different scenarios.
type MockQuerier struct {
	// Fields to control what the mock returns
	CreateUserFunc    func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserByIDFunc   func(ctx context.Context, userID uuid.UUID) (database.User, error)
	ListUsersFunc     func(ctx context.Context) ([]database.User, error)
	UpdateUserFunc    func(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
	DeleteUserFunc    func(ctx context.Context, userID uuid.UUID) error
	EmailExistsFunc   func(ctx context.Context, email string) (bool, error)
	UserExistsFunc    func(ctx context.Context, userID uuid.UUID) (bool, error)
	GetUserByEmailFunc func(ctx context.Context, email string) (database.User, error)
	ListUsersByStatusFunc func(ctx context.Context, status string) ([]database.User, error)
}

// Implement all methods required by the Querier interface
// Each method calls the corresponding function field if set

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

// Verify MockQuerier implements Querier interface at compile time
var _ database.Querier = (*MockQuerier)(nil)

// ============================================================================
// Helper Functions
// ============================================================================

// createMockUser creates a mock database user for testing
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
// CreateUser Tests
// ============================================================================

func TestUserService_CreateUser_Success(t *testing.T) {
	// Arrange: Set up the mock
	mockQuerier := &MockQuerier{
		// Email doesn't exist yet
		EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil // Email is available
		},
		// CreateUser succeeds
		CreateUserFunc: func(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
			return createMockUser(uuid.New(), arg.FirstName, arg.LastName, arg.Email), nil
		},
	}

	// Create service with mock (nil for pool since we're mocking)
	svc := service.NewUserService(nil, database.New(mockQuerier))

	// Act: Call the method being tested
	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}
	result, err := svc.CreateUser(context.Background(), req)

	// Assert: Check the results
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.FirstName != "John" {
		t.Errorf("Expected FirstName 'John', got '%s'", result.FirstName)
	}
	if result.Email != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got '%s'", result.Email)
	}
}

func TestUserService_CreateUser_EmailAlreadyExists(t *testing.T) {
	mockQuerier := &MockQuerier{
		EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Email already exists!
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "existing@example.com",
	}
	result, err := svc.CreateUser(context.Background(), req)

	// Should return error
	if err == nil {
		t.Fatal("Expected error for duplicate email")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}

	// Check it's a ConflictError (409)
	appErr, ok := err.(*models.AppError)
	if !ok {
		t.Fatal("Expected AppError type")
	}
	if appErr.StatusCode != 409 {
		t.Errorf("Expected status 409, got %d", appErr.StatusCode)
	}
}

func TestUserService_CreateUser_DatabaseError(t *testing.T) {
	mockQuerier := &MockQuerier{
		EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		CreateUserFunc: func(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
			return database.User{}, errors.New("database connection failed")
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}
	_, err := svc.CreateUser(context.Background(), req)

	// Should return internal server error
	if err == nil {
		t.Fatal("Expected error for database failure")
	}

	appErr, ok := err.(*models.AppError)
	if !ok {
		t.Fatal("Expected AppError type")
	}
	if appErr.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", appErr.StatusCode)
	}
}

// ============================================================================
// GetUserByID Tests
// ============================================================================

func TestUserService_GetUserByID_Success(t *testing.T) {
	userID := uuid.New()
	mockUser := createMockUser(userID, "Jane", "Doe", "jane@example.com")

	mockQuerier := &MockQuerier{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
			if id == userID {
				return mockUser, nil
			}
			return database.User{}, pgx.ErrNoRows
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	result, err := svc.GetUserByID(context.Background(), userID.String())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.FirstName != "Jane" {
		t.Errorf("Expected 'Jane', got '%s'", result.FirstName)
	}
}

func TestUserService_GetUserByID_InvalidUUID(t *testing.T) {
	mockQuerier := &MockQuerier{}
	svc := service.NewUserService(nil, database.New(mockQuerier))

	// Pass invalid UUID format
	_, err := svc.GetUserByID(context.Background(), "not-a-valid-uuid")

	if err == nil {
		t.Fatal("Expected error for invalid UUID")
	}

	appErr, ok := err.(*models.AppError)
	if !ok {
		t.Fatal("Expected AppError type")
	}
	if appErr.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", appErr.StatusCode)
	}
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	mockQuerier := &MockQuerier{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
			return database.User{}, pgx.ErrNoRows // User not found
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	_, err := svc.GetUserByID(context.Background(), uuid.New().String())

	if err == nil {
		t.Fatal("Expected error for user not found")
	}

	appErr, ok := err.(*models.AppError)
	if !ok {
		t.Fatal("Expected AppError type")
	}
	if appErr.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", appErr.StatusCode)
	}
}

// ============================================================================
// ListUsers Tests
// ============================================================================

func TestUserService_ListUsers_Success(t *testing.T) {
	mockUsers := []database.User{
		createMockUser(uuid.New(), "John", "Doe", "john@example.com"),
		createMockUser(uuid.New(), "Jane", "Smith", "jane@example.com"),
	}

	mockQuerier := &MockQuerier{
		ListUsersFunc: func(ctx context.Context) ([]database.User, error) {
			return mockUsers, nil
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	result, err := svc.ListUsers(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Expected 2 users, got %d", result.Total)
	}
	if len(result.Users) != 2 {
		t.Errorf("Expected 2 users in slice, got %d", len(result.Users))
	}
}

func TestUserService_ListUsers_Empty(t *testing.T) {
	mockQuerier := &MockQuerier{
		ListUsersFunc: func(ctx context.Context) ([]database.User, error) {
			return []database.User{}, nil // Empty list
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	result, err := svc.ListUsers(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("Expected 0 users, got %d", result.Total)
	}
}

// ============================================================================
// DeleteUser Tests
// ============================================================================

func TestUserService_DeleteUser_Success(t *testing.T) {
	userID := uuid.New()

	mockQuerier := &MockQuerier{
		UserExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil // User exists
		},
		DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil // Delete succeeds
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	err := svc.DeleteUser(context.Background(), userID.String())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	mockQuerier := &MockQuerier{
		UserExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, nil // User doesn't exist
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	err := svc.DeleteUser(context.Background(), uuid.New().String())

	if err == nil {
		t.Fatal("Expected error for user not found")
	}

	appErr, ok := err.(*models.AppError)
	if !ok {
		t.Fatal("Expected AppError type")
	}
	if appErr.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", appErr.StatusCode)
	}
}

// ============================================================================
// UpdateUser Tests
// ============================================================================

func TestUserService_UpdateUser_Success(t *testing.T) {
	userID := uuid.New()
	newFirstName := "UpdatedJohn"

	mockQuerier := &MockQuerier{
		UserExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
		UpdateUserFunc: func(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
			return database.User{
				UserID:    arg.UserID,
				FirstName: newFirstName,
				LastName:  "Doe",
				Email:     "john@example.com",
				Status:    "Active",
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	req := models.UpdateUserRequest{
		FirstName: &newFirstName,
	}
	result, err := svc.UpdateUser(context.Background(), userID.String(), req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.FirstName != newFirstName {
		t.Errorf("Expected '%s', got '%s'", newFirstName, result.FirstName)
	}
}

func TestUserService_UpdateUser_EmailConflict(t *testing.T) {
	userID := uuid.New()
	existingEmail := "existing@example.com"

	mockQuerier := &MockQuerier{
		UserExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
		EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Email already exists
		},
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (database.User, error) {
			return database.User{
				UserID: id,
				Email:  "current@example.com", // Different from new email
			}, nil
		},
	}

	svc := service.NewUserService(nil, database.New(mockQuerier))

	req := models.UpdateUserRequest{
		Email: &existingEmail,
	}
	_, err := svc.UpdateUser(context.Background(), userID.String(), req)

	if err == nil {
		t.Fatal("Expected error for email conflict")
	}

	appErr, ok := err.(*models.AppError)
	if !ok {
		t.Fatal("Expected AppError type")
	}
	if appErr.StatusCode != 409 {
		t.Errorf("Expected status 409, got %d", appErr.StatusCode)
	}
}
