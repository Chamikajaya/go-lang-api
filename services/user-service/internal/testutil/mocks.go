package testutil

import (
	"context"
	"time"

	database "user-service/db/sqlc"
	"user-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// =============================================================================
// Shared Test Fixtures
// =============================================================================

// TestUserID is a fixed UUID used across tests for deterministic assertions.
var TestUserID = uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

// NewDBUser builds a database.User with sensible sample values.
func NewDBUser() database.User {
	now := time.Now().UTC()
	return database.User{
		UserID:    TestUserID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     pgtype.Text{String: "+1234567890", Valid: true},
		Age:       pgtype.Int4{Int32: 30, Valid: true},
		Status:    "Active",
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}
}

// StrPtr returns a pointer to the given string. Useful for optional fields in update requests.
func StrPtr(s string) *string { return &s }

// =============================================================================
// MockQuerier — implements database.Querier
// =============================================================================

// MockQuerier implements database.Querier to simulate DB operations in tests.
// Set function fields to control behavior; unset fields will panic if called.
type MockQuerier struct {
	CreateUserFn      func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	DeleteUserFn      func(ctx context.Context, userID uuid.UUID) error
	EmailExistsFn     func(ctx context.Context, email string) (bool, error)
	GetUserByEmailFn  func(ctx context.Context, email string) (database.User, error)
	GetUserByIDFn     func(ctx context.Context, userID uuid.UUID) (database.User, error)
	ListUsersFn       func(ctx context.Context) ([]database.User, error)
	ListUsersByStatFn func(ctx context.Context, status string) ([]database.User, error)
	UpdateUserFn      func(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
	UserExistsFn      func(ctx context.Context, userID uuid.UUID) (bool, error)
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return m.CreateUserFn(ctx, arg)
}
func (m *MockQuerier) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return m.DeleteUserFn(ctx, userID)
}
func (m *MockQuerier) EmailExists(ctx context.Context, email string) (bool, error) {
	return m.EmailExistsFn(ctx, email)
}
func (m *MockQuerier) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return m.GetUserByEmailFn(ctx, email)
}
func (m *MockQuerier) GetUserByID(ctx context.Context, userID uuid.UUID) (database.User, error) {
	return m.GetUserByIDFn(ctx, userID)
}
func (m *MockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	return m.ListUsersFn(ctx)
}
func (m *MockQuerier) ListUsersByStatus(ctx context.Context, status string) ([]database.User, error) {
	return m.ListUsersByStatFn(ctx, status)
}
func (m *MockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return m.UpdateUserFn(ctx, arg)
}
func (m *MockQuerier) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.UserExistsFn(ctx, userID)
}

// =============================================================================
// MockPublisher — implements publisher.EventPublisherInterface
// =============================================================================

// MockPublisher implements publisher.EventPublisherInterface.
// Boolean flags track which events were published; set Err to simulate publish failures.
type MockPublisher struct {
	PublishedCreated bool
	PublishedUpdated bool
	PublishedDeleted bool
	Err              error
}

func (m *MockPublisher) PublishUserCreated(_ context.Context, _ string, _ *models.UserResponse) error {
	m.PublishedCreated = true
	return m.Err
}
func (m *MockPublisher) PublishUserUpdated(_ context.Context, _ string, _ *models.UserResponse) error {
	m.PublishedUpdated = true
	return m.Err
}
func (m *MockPublisher) PublishUserDeleted(_ context.Context, _ string, _ string) error {
	m.PublishedDeleted = true
	return m.Err
}
