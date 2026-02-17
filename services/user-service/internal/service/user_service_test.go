package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	database "user-service/db/sqlc"
	"user-service/internal/models"
	"user-service/internal/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// =============================================================================
// Helpers (service-layer specific)
// =============================================================================

// newService creates a UserService wired with the given mock querier and publisher.
func newService(q *testutil.MockQuerier, p *testutil.MockPublisher) *UserService {
	return NewUserService(nil, q, p) // pool is nil because the service layer doesn't use it directly.
}

// assertAppError checks the returned error is an *AppError with the expected HTTP status code.
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
	dbUser := testutil.NewDBUser()
	q := &testutil.MockQuerier{
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) {
			return false, nil // mock false → email not taken
		},
		CreateUserFn: func(_ context.Context, arg database.CreateUserParams) (database.User, error) {
			// Verify the params were mapped correctly
			if arg.FirstName != "John" || arg.Email != "john@example.com" {
				t.Errorf("unexpected create params: %+v", arg)
			}
			return dbUser, nil
		},
	}
	pub := &testutil.MockPublisher{}
	svc := newService(q, pub)

	req := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	// Act
	// calling the actual service method want to test
	resp, err := svc.CreateUser(context.Background(), req, "actor-1")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.ID != testutil.TestUserID {
		t.Errorf("expected user ID %s, got %s", testutil.TestUserID, resp.ID)
	}
	if resp.FirstName != "John" {
		t.Errorf("expected firstName John, got %s", resp.FirstName)
	}

	// veryfying that the event was published
	if !pub.PublishedCreated {
		t.Error("expected user.created event to be published")
	}
}

func TestCreateUser_DefaultsToActiveStatus(t *testing.T) {
	// Arrange: no status provided — should default to Active
	q := &testutil.MockQuerier{
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		CreateUserFn: func(_ context.Context, arg database.CreateUserParams) (database.User, error) {
			if arg.Status != "Active" {
				t.Errorf("expected default status 'Active', got %q", arg.Status)
			}
			return testutil.NewDBUser(), nil
		},
	}
	svc := newService(q, &testutil.MockPublisher{})

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
	q := &testutil.MockQuerier{
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) {
			return true, nil // email taken - mocking true to simulate conflict
		},
	}
	pub := &testutil.MockPublisher{}
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

	// event should NOT be published on conflict
	if pub.PublishedCreated {
		t.Error("should NOT publish event on conflict")
	}
}

// =============================================================================
// GetUserByID Tests
// =============================================================================

func TestGetUserByID_Success(t *testing.T) {
	dbUser := testutil.NewDBUser()
	q := &testutil.MockQuerier{
		GetUserByIDFn: func(_ context.Context, id uuid.UUID) (database.User, error) {
			if id != testutil.TestUserID {
				t.Errorf("unexpected user ID: %s", id)
			}
			return dbUser, nil
		},
	}
	svc := newService(q, &testutil.MockPublisher{})

	resp, err := svc.GetUserByID(context.Background(), testutil.TestUserID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", resp.Email)
	}
}

func TestGetUserByID_InvalidUUID(t *testing.T) {
	svc := newService(&testutil.MockQuerier{}, &testutil.MockPublisher{})

	_, err := svc.GetUserByID(context.Background(), "not-a-uuid")

	// Should return 400 Bad Request
	assertAppError(t, err, http.StatusBadRequest)
}

func TestGetUserByID_NotFound(t *testing.T) {
	q := &testutil.MockQuerier{
		GetUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return database.User{}, pgx.ErrNoRows
		},
	}
	svc := newService(q, &testutil.MockPublisher{})

	_, err := svc.GetUserByID(context.Background(), testutil.TestUserID.String())

	// Should return 404 Not Found
	assertAppError(t, err, http.StatusNotFound)
}

// =============================================================================
// ListUsers Tests
// =============================================================================

func TestListUsers_Success(t *testing.T) {
	user1 := testutil.NewDBUser()
	user2 := testutil.NewDBUser()
	user2.UserID = uuid.New()
	user2.Email = "jane@example.com"

	q := &testutil.MockQuerier{
		ListUsersFn: func(_ context.Context) ([]database.User, error) {
			return []database.User{user1, user2}, nil
		},
	}
	svc := newService(q, &testutil.MockPublisher{})

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
	q := &testutil.MockQuerier{
		ListUsersFn: func(_ context.Context) ([]database.User, error) {
			return []database.User{}, nil
		},
	}
	svc := newService(q, &testutil.MockPublisher{})

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
	dbUser := testutil.NewDBUser()
	dbUser.FirstName = "Jane" // updated name

	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		UpdateUserFn: func(_ context.Context, _ database.UpdateUserParams) (database.User, error) {
			return dbUser, nil
		},
	}
	pub := &testutil.MockPublisher{}
	svc := newService(q, pub)

	newName := "Jane"
	req := models.UpdateUserRequest{FirstName: &newName}

	resp, err := svc.UpdateUser(context.Background(), testutil.TestUserID.String(), req, "actor-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FirstName != "Jane" {
		t.Errorf("expected updated firstName Jane, got %s", resp.FirstName)
	}
	if !pub.PublishedUpdated {
		t.Error("expected user.updated event to be published")
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
	}
	svc := newService(q, &testutil.MockPublisher{})

	newName := "Jane"
	req := models.UpdateUserRequest{FirstName: &newName}

	_, err := svc.UpdateUser(context.Background(), testutil.TestUserID.String(), req, "actor-1")

	assertAppError(t, err, http.StatusNotFound)
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	// Existing user has email "john@example.com", trying to update to "taken@example.com"
	// which already belongs to a different user.
	existingUser := testutil.NewDBUser()

	q := &testutil.MockQuerier{
		UserExistsFn:  func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil }, // email taken
		GetUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return existingUser, nil // current user's email is "john@example.com"
		},
	}
	pub := &testutil.MockPublisher{}
	svc := newService(q, pub)

	newEmail := "taken@example.com" // different from current email → conflict
	req := models.UpdateUserRequest{Email: &newEmail}

	_, err := svc.UpdateUser(context.Background(), testutil.TestUserID.String(), req, "actor-1")

	assertAppError(t, err, http.StatusConflict)
	if pub.PublishedUpdated {
		t.Error("should NOT publish event on conflict")
	}
}

func TestUpdateUser_SameEmailAllowed(t *testing.T) {
	// If the user updates with their own current email, it should NOT conflict.
	existingUser := testutil.NewDBUser()

	q := &testutil.MockQuerier{
		UserExistsFn:  func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil }, // email "exists"
		GetUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return existingUser, nil
		},
		UpdateUserFn: func(_ context.Context, _ database.UpdateUserParams) (database.User, error) {
			return existingUser, nil
		},
	}
	svc := newService(q, &testutil.MockPublisher{})

	sameEmail := "john@example.com" // same as existingUser.Email
	req := models.UpdateUserRequest{Email: &sameEmail}

	_, err := svc.UpdateUser(context.Background(), testutil.TestUserID.String(), req, "actor-1")
	if err != nil {
		t.Fatalf("expected no error when updating with own email, got: %v", err)
	}
}

// =============================================================================
// DeleteUser Tests
// =============================================================================

func TestDeleteUser_Success(t *testing.T) {
	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		DeleteUserFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	pub := &testutil.MockPublisher{}
	svc := newService(q, pub)

	err := svc.DeleteUser(context.Background(), testutil.TestUserID.String(), "actor-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pub.PublishedDeleted {
		t.Error("expected user.deleted event to be published")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
	}
	pub := &testutil.MockPublisher{}
	svc := newService(q, pub)

	err := svc.DeleteUser(context.Background(), testutil.TestUserID.String(), "actor-1")

	assertAppError(t, err, http.StatusNotFound)
	if pub.PublishedDeleted {
		t.Error("should NOT publish event when user not found")
	}
}

func TestDeleteUser_InvalidUUID(t *testing.T) {
	svc := newService(&testutil.MockQuerier{}, &testutil.MockPublisher{})

	err := svc.DeleteUser(context.Background(), "bad-id", "actor-1")

	assertAppError(t, err, http.StatusBadRequest)
}
