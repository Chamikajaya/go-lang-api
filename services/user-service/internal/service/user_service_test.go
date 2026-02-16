package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	database "user-service/db/sqlc"
	"user-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// =============================================================================
// Mocks
// =============================================================================

// mockQuerier implements database.Querier to simulate DB operations in tests.
// Each field is a function that the test can override to control behaviour.
type mockQuerier struct {
	createUserFn      func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	deleteUserFn      func(ctx context.Context, userID uuid.UUID) error
	emailExistsFn     func(ctx context.Context, email string) (bool, error)
	getUserByEmailFn  func(ctx context.Context, email string) (database.User, error)
	getUserByIDFn     func(ctx context.Context, userID uuid.UUID) (database.User, error)
	listUsersFn       func(ctx context.Context) ([]database.User, error)
	listUsersByStatFn func(ctx context.Context, status string) ([]database.User, error)
	updateUserFn      func(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
	userExistsFn      func(ctx context.Context, userID uuid.UUID) (bool, error)
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return m.createUserFn(ctx, arg)
}
func (m *mockQuerier) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return m.deleteUserFn(ctx, userID)
}
func (m *mockQuerier) EmailExists(ctx context.Context, email string) (bool, error) {
	return m.emailExistsFn(ctx, email)
}
func (m *mockQuerier) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return m.getUserByEmailFn(ctx, email)
}
func (m *mockQuerier) GetUserByID(ctx context.Context, userID uuid.UUID) (database.User, error) {
	return m.getUserByIDFn(ctx, userID)
}
func (m *mockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	return m.listUsersFn(ctx)
}
func (m *mockQuerier) ListUsersByStatus(ctx context.Context, status string) ([]database.User, error) {
	return m.listUsersByStatFn(ctx, status)
}
func (m *mockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return m.updateUserFn(ctx, arg)
}
func (m *mockQuerier) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.userExistsFn(ctx, userID)
}

// mockPublisher implements publisher.EventPublisherInterface.
// It records calls so tests can assert that events were published.
type mockPublisher struct {
	publishedCreated bool
	publishedUpdated bool
	publishedDeleted bool
	err              error // if set, Publish* methods return this error
}

func (m *mockPublisher) PublishUserCreated(_ context.Context, _ string, _ *models.UserResponse) error {
	m.publishedCreated = true
	return m.err
}
func (m *mockPublisher) PublishUserUpdated(_ context.Context, _ string, _ *models.UserResponse) error {
	m.publishedUpdated = true
	return m.err
}
func (m *mockPublisher) PublishUserDeleted(_ context.Context, _ string, _ string) error {
	m.publishedDeleted = true
	return m.err
}

// =============================================================================
// Helpers
// =============================================================================

// newTestUserID returns a fixed UUID for deterministic tests.
var testUserID = uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

// newDBUser builds a database.User with sensible defaults.
func newDBUser() database.User {
	now := time.Now().UTC()
	return database.User{
		UserID:    testUserID,
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

// newService creates a UserService wired with the given mock querier and publisher.
// pool is nil because the service layer doesn't use it directly (queries do).
func newService(q *mockQuerier, p *mockPublisher) *UserService {
	return NewUserService(nil, q, p)
}

// assertAppError is a helper that checks the returned error is an *AppError
// with the expected HTTP status code.
func assertAppError(t *testing.T, err error, expectedStatus int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with status %d, got nil", expectedStatus)
	}
	var appErr *models.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *AppError, got %T: %v", err, err)
	}
	if appErr.StatusCode != expectedStatus {
		t.Errorf("expected status %d, got %d (message: %s)", expectedStatus, appErr.StatusCode, appErr.Message)
	}
}

// =============================================================================
// CreateUser Tests
// =============================================================================

func TestCreateUser_Success(t *testing.T) {
	// Arrange: email does not exist, DB creates user successfully
	dbUser := newDBUser()
	q := &mockQuerier{
		emailExistsFn: func(_ context.Context, _ string) (bool, error) {
			return false, nil
		},
		createUserFn: func(_ context.Context, arg database.CreateUserParams) (database.User, error) {
			// Verify the params were mapped correctly
			if arg.FirstName != "John" || arg.Email != "john@example.com" {
				t.Errorf("unexpected create params: %+v", arg)
			}
			return dbUser, nil
		},
	}
	pub := &mockPublisher{}
	svc := newService(q, pub)

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	// Act
	resp, err := svc.CreateUser(context.Background(), req, "actor-1")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.ID != testUserID {
		t.Errorf("expected user ID %s, got %s", testUserID, resp.ID)
	}
	if resp.FirstName != "John" {
		t.Errorf("expected firstName John, got %s", resp.FirstName)
	}
	if !pub.publishedCreated {
		t.Error("expected user.created event to be published")
	}
}

func TestCreateUser_DefaultsToActiveStatus(t *testing.T) {
	// Arrange: no status provided — should default to Active
	q := &mockQuerier{
		emailExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		createUserFn: func(_ context.Context, arg database.CreateUserParams) (database.User, error) {
			if arg.Status != "Active" {
				t.Errorf("expected default status 'Active', got %q", arg.Status)
			}
			return newDBUser(), nil
		},
	}
	svc := newService(q, &mockPublisher{})

	req := models.CreateUserRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		// Status intentionally left empty
	}

	_, err := svc.CreateUser(context.Background(), req, "actor-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateUser_EmailConflict(t *testing.T) {
	// Arrange: email already exists
	q := &mockQuerier{
		emailExistsFn: func(_ context.Context, _ string) (bool, error) {
			return true, nil // email taken
		},
	}
	pub := &mockPublisher{}
	svc := newService(q, pub)

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "taken@example.com",
	}

	// Act
	resp, err := svc.CreateUser(context.Background(), req, "actor-1")

	// Assert: should return 409 Conflict
	if resp != nil {
		t.Errorf("expected nil response on conflict, got %+v", resp)
	}
	assertAppError(t, err, http.StatusConflict)
	if pub.publishedCreated {
		t.Error("should NOT publish event on conflict")
	}
}

// =============================================================================
// GetUserByID Tests
// =============================================================================

func TestGetUserByID_Success(t *testing.T) {
	dbUser := newDBUser()
	q := &mockQuerier{
		getUserByIDFn: func(_ context.Context, id uuid.UUID) (database.User, error) {
			if id != testUserID {
				t.Errorf("unexpected user ID: %s", id)
			}
			return dbUser, nil
		},
	}
	svc := newService(q, &mockPublisher{})

	resp, err := svc.GetUserByID(context.Background(), testUserID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", resp.Email)
	}
}

func TestGetUserByID_InvalidUUID(t *testing.T) {
	svc := newService(&mockQuerier{}, &mockPublisher{})

	_, err := svc.GetUserByID(context.Background(), "not-a-uuid")

	// Should return 400 Bad Request
	assertAppError(t, err, http.StatusBadRequest)
}

func TestGetUserByID_NotFound(t *testing.T) {
	q := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return database.User{}, pgx.ErrNoRows
		},
	}
	svc := newService(q, &mockPublisher{})

	_, err := svc.GetUserByID(context.Background(), testUserID.String())

	// Should return 404 Not Found
	assertAppError(t, err, http.StatusNotFound)
}

// =============================================================================
// ListUsers Tests
// =============================================================================

func TestListUsers_Success(t *testing.T) {
	user1 := newDBUser()
	user2 := newDBUser()
	user2.UserID = uuid.New()
	user2.Email = "jane@example.com"

	q := &mockQuerier{
		listUsersFn: func(_ context.Context) ([]database.User, error) {
			return []database.User{user1, user2}, nil
		},
	}
	svc := newService(q, &mockPublisher{})

	resp, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(resp.Users))
	}
}

func TestListUsers_Empty(t *testing.T) {
	q := &mockQuerier{
		listUsersFn: func(_ context.Context) ([]database.User, error) {
			return []database.User{}, nil
		},
	}
	svc := newService(q, &mockPublisher{})

	resp, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Total)
	}
}

// =============================================================================
// UpdateUser Tests
// =============================================================================

func TestUpdateUser_Success(t *testing.T) {
	dbUser := newDBUser()
	dbUser.FirstName = "Jane" // updated name

	q := &mockQuerier{
		userExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		updateUserFn: func(_ context.Context, _ database.UpdateUserParams) (database.User, error) {
			return dbUser, nil
		},
	}
	pub := &mockPublisher{}
	svc := newService(q, pub)

	newName := "Jane"
	req := models.UpdateUserRequest{FirstName: &newName}

	resp, err := svc.UpdateUser(context.Background(), testUserID.String(), req, "actor-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FirstName != "Jane" {
		t.Errorf("expected updated firstName Jane, got %s", resp.FirstName)
	}
	if !pub.publishedUpdated {
		t.Error("expected user.updated event to be published")
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	q := &mockQuerier{
		userExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
	}
	svc := newService(q, &mockPublisher{})

	newName := "Jane"
	req := models.UpdateUserRequest{FirstName: &newName}

	_, err := svc.UpdateUser(context.Background(), testUserID.String(), req, "actor-1")

	assertAppError(t, err, http.StatusNotFound)
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	// Existing user has email "john@example.com", trying to update to "taken@example.com"
	// which already belongs to a different user.
	existingUser := newDBUser()

	q := &mockQuerier{
		userExistsFn:  func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		emailExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil }, // email taken
		getUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return existingUser, nil // current user's email is "john@example.com"
		},
	}
	pub := &mockPublisher{}
	svc := newService(q, pub)

	newEmail := "taken@example.com" // different from current email → conflict
	req := models.UpdateUserRequest{Email: &newEmail}

	_, err := svc.UpdateUser(context.Background(), testUserID.String(), req, "actor-1")

	assertAppError(t, err, http.StatusConflict)
	if pub.publishedUpdated {
		t.Error("should NOT publish event on conflict")
	}
}

func TestUpdateUser_SameEmailAllowed(t *testing.T) {
	// If the user updates with their own current email, it should NOT conflict.
	existingUser := newDBUser()

	q := &mockQuerier{
		userExistsFn:  func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		emailExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil }, // email "exists"
		getUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return existingUser, nil
		},
		updateUserFn: func(_ context.Context, _ database.UpdateUserParams) (database.User, error) {
			return existingUser, nil
		},
	}
	svc := newService(q, &mockPublisher{})

	sameEmail := "john@example.com" // same as existingUser.Email
	req := models.UpdateUserRequest{Email: &sameEmail}

	_, err := svc.UpdateUser(context.Background(), testUserID.String(), req, "actor-1")
	if err != nil {
		t.Fatalf("expected no error when updating with own email, got: %v", err)
	}
}

// =============================================================================
// DeleteUser Tests
// =============================================================================

func TestDeleteUser_Success(t *testing.T) {
	q := &mockQuerier{
		userExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		deleteUserFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	pub := &mockPublisher{}
	svc := newService(q, pub)

	err := svc.DeleteUser(context.Background(), testUserID.String(), "actor-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pub.publishedDeleted {
		t.Error("expected user.deleted event to be published")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	q := &mockQuerier{
		userExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
	}
	pub := &mockPublisher{}
	svc := newService(q, pub)

	err := svc.DeleteUser(context.Background(), testUserID.String(), "actor-1")

	assertAppError(t, err, http.StatusNotFound)
	if pub.publishedDeleted {
		t.Error("should NOT publish event when user not found")
	}
}

func TestDeleteUser_InvalidUUID(t *testing.T) {
	svc := newService(&mockQuerier{}, &mockPublisher{})

	err := svc.DeleteUser(context.Background(), "bad-id", "actor-1")

	assertAppError(t, err, http.StatusBadRequest)
}
