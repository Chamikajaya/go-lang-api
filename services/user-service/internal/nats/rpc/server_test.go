package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	database "user-service/db/sqlc"
	"user-service/internal/models"
	"user-service/internal/service"
	"user-service/internal/testutil"
	"user-service/internal/validator"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// =============================================================================
// RPC-layer Helpers
// =============================================================================

func startTestNATS(t *testing.T) (*server.Server, *nats.Conn) {
	t.Helper()
	opts := &server.Options{
		Host: "127.0.0.1",
		Port: -1,
	}
	ns, err := server.NewServer(opts)
	if err != nil {
		t.Fatalf("failed to create test NATS server: %v", err)
	}
	ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		t.Fatal("NATS server not ready")
	}
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		ns.Shutdown()
		t.Fatalf("failed to connect to test NATS: %v", err)
	}
	return ns, nc
}

func rpcRequest(t *testing.T, nc *nats.Conn, subject string, payload interface{}) *RPCResponse {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	msg, err := nc.Request(subject, data, 5*time.Second)
	if err != nil {
		t.Fatalf("NATS request to %s failed: %v", subject, err)
	}
	var resp RPCResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		t.Fatalf("failed to unmarshal RPC response: %v", err)
	}
	return &resp
}

func setupRPCServer(t *testing.T, q *testutil.MockQuerier) (*Server, *nats.Conn, func()) {
	t.Helper()
	ns, nc := startTestNATS(t)
	svc := service.NewUserService(nil, q, &testutil.MockPublisher{})
	v := validator.NewValidator()
	rpcSrv := NewServer(nc, svc, v)
	if err := rpcSrv.Start(); err != nil {
		nc.Close()
		ns.Shutdown()
		t.Fatalf("failed to start RPC server: %v", err)
	}
	cleanup := func() {
		_ = rpcSrv.Stop()
		nc.Close()
		ns.Shutdown()
	}
	return rpcSrv, nc, cleanup
}

func assertRPCSuccess(t *testing.T, resp *RPCResponse) {
	t.Helper()
	if !resp.Success {
		t.Fatalf("expected success, got error: %+v", resp.Error)
	}
}

func assertRPCError(t *testing.T, resp *RPCResponse, expectedCode int) {
	t.Helper()
	if resp.Success {
		t.Fatal("expected failure, got success")
	}
	if resp.Error.Code != expectedCode {
		t.Errorf("expected status %d, got %d", expectedCode, resp.Error.Code)
	}
}

func sendRawAndDecode(t *testing.T, nc *nats.Conn, subject string, raw []byte) *RPCResponse {
	t.Helper()
	msg, err := nc.Request(subject, raw, 5*time.Second)
	if err != nil {
		t.Fatalf("NATS request failed: %v", err)
	}
	var resp RPCResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return &resp
}

// =============================================================================
// handleCreateUser Tests
// =============================================================================

func TestRPC_CreateUser_Success(t *testing.T) {
	dbUser := testutil.NewDBUser()
	q := &testutil.MockQuerier{
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		CreateUserFn: func(_ context.Context, _ database.CreateUserParams) (database.User, error) {
			return dbUser, nil
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := CreateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		Data: models.CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}
	resp := rpcRequest(t, nc, SubjectUserCreate, req)
	assertRPCSuccess(t, resp)
}

func TestRPC_CreateUser_MalformedJSON(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	resp := sendRawAndDecode(t, nc, SubjectUserCreate, []byte("{invalid"))
	assertRPCError(t, resp, http.StatusBadRequest)
}

func TestRPC_CreateUser_ValidationError(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := CreateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		Data:       models.CreateUserRequest{},
	}
	resp := rpcRequest(t, nc, SubjectUserCreate, req)
	assertRPCError(t, resp, http.StatusBadRequest)
	if resp.Error.Message != "Validation failed" {
		t.Errorf("expected Validation failed message, got %q", resp.Error.Message)
	}
}

func TestRPC_CreateUser_EmailConflict(t *testing.T) {
	q := &testutil.MockQuerier{
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := CreateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		Data: models.CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "taken@example.com",
		},
	}
	resp := rpcRequest(t, nc, SubjectUserCreate, req)
	assertRPCError(t, resp, http.StatusConflict)
}

// =============================================================================
// handleGetUser Tests
// =============================================================================

func TestRPC_GetUser_Success(t *testing.T) {
	dbUser := testutil.NewDBUser()
	q := &testutil.MockQuerier{
		GetUserByIDFn: func(_ context.Context, id uuid.UUID) (database.User, error) {
			if id != testutil.TestUserID {
				t.Errorf("unexpected user ID: %s", id)
			}
			return dbUser, nil
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := GetUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
	}
	resp := rpcRequest(t, nc, SubjectUserGet, req)
	assertRPCSuccess(t, resp)
}

func TestRPC_GetUser_InvalidUUID(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := GetUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     "not-a-uuid",
	}
	resp := rpcRequest(t, nc, SubjectUserGet, req)
	assertRPCError(t, resp, http.StatusBadRequest)
}

func TestRPC_GetUser_NotFound(t *testing.T) {
	q := &testutil.MockQuerier{
		GetUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return database.User{}, pgx.ErrNoRows
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := GetUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
	}
	resp := rpcRequest(t, nc, SubjectUserGet, req)
	assertRPCError(t, resp, http.StatusNotFound)
}

func TestRPC_GetUser_MalformedJSON(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	resp := sendRawAndDecode(t, nc, SubjectUserGet, []byte("not json"))
	assertRPCError(t, resp, http.StatusBadRequest)
}

// =============================================================================
// handleListUsers Tests
// =============================================================================

func TestRPC_ListUsers_Success(t *testing.T) {
	user1 := testutil.NewDBUser()
	user2 := testutil.NewDBUser()
	user2.UserID = uuid.New()
	user2.Email = "jane@example.com"

	q := &testutil.MockQuerier{
		ListUsersFn: func(_ context.Context) ([]database.User, error) {
			return []database.User{user1, user2}, nil
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := ListUsersRPCRequest{RPCRequest: RPCRequest{ActorID: "actor-1"}}
	resp := rpcRequest(t, nc, SubjectUserList, req)
	assertRPCSuccess(t, resp)
}

func TestRPC_ListUsers_Empty(t *testing.T) {
	q := &testutil.MockQuerier{
		ListUsersFn: func(_ context.Context) ([]database.User, error) {
			return []database.User{}, nil
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := ListUsersRPCRequest{RPCRequest: RPCRequest{ActorID: "actor-1"}}
	resp := rpcRequest(t, nc, SubjectUserList, req)
	assertRPCSuccess(t, resp)
}

func TestRPC_ListUsers_InternalError(t *testing.T) {
	q := &testutil.MockQuerier{
		ListUsersFn: func(_ context.Context) ([]database.User, error) {
			return nil, models.NewInternalServerError("db down", nil)
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := ListUsersRPCRequest{RPCRequest: RPCRequest{ActorID: "actor-1"}}
	resp := rpcRequest(t, nc, SubjectUserList, req)
	assertRPCError(t, resp, http.StatusInternalServerError)
}

// =============================================================================
// handleUpdateUser Tests
// =============================================================================

func TestRPC_UpdateUser_Success(t *testing.T) {
	dbUser := testutil.NewDBUser()
	dbUser.FirstName = "Jane"

	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		UpdateUserFn: func(_ context.Context, _ database.UpdateUserParams) (database.User, error) {
			return dbUser, nil
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := UpdateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
		Data:       models.UpdateUserRequest{FirstName: testutil.StrPtr("Jane")},
	}
	resp := rpcRequest(t, nc, SubjectUserUpdate, req)
	assertRPCSuccess(t, resp)
}

func TestRPC_UpdateUser_MalformedJSON(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	resp := sendRawAndDecode(t, nc, SubjectUserUpdate, []byte("broken"))
	assertRPCError(t, resp, http.StatusBadRequest)
}

func TestRPC_UpdateUser_ValidationError(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	shortName := "A"
	req := UpdateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
		Data:       models.UpdateUserRequest{FirstName: &shortName},
	}
	resp := rpcRequest(t, nc, SubjectUserUpdate, req)
	assertRPCError(t, resp, http.StatusBadRequest)
	if resp.Error.Details == nil {
		t.Error("expected validation details in error response")
	}
}

func TestRPC_UpdateUser_NotFound(t *testing.T) {
	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := UpdateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
		Data:       models.UpdateUserRequest{FirstName: testutil.StrPtr("Jane")},
	}
	resp := rpcRequest(t, nc, SubjectUserUpdate, req)
	assertRPCError(t, resp, http.StatusNotFound)
}

func TestRPC_UpdateUser_EmailConflict(t *testing.T) {
	existingUser := testutil.NewDBUser()
	q := &testutil.MockQuerier{
		UserExistsFn:  func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		EmailExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil },
		GetUserByIDFn: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return existingUser, nil
		},
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := UpdateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
		Data:       models.UpdateUserRequest{Email: testutil.StrPtr("taken@example.com")},
	}
	resp := rpcRequest(t, nc, SubjectUserUpdate, req)
	assertRPCError(t, resp, http.StatusConflict)
}

func TestRPC_UpdateUser_InvalidUUID(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := UpdateUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     "not-valid",
		Data:       models.UpdateUserRequest{FirstName: testutil.StrPtr("Jane")},
	}
	resp := rpcRequest(t, nc, SubjectUserUpdate, req)
	assertRPCError(t, resp, http.StatusBadRequest)
}

// =============================================================================
// handleDeleteUser Tests
// =============================================================================

func TestRPC_DeleteUser_Success(t *testing.T) {
	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
		DeleteUserFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := DeleteUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
	}
	resp := rpcRequest(t, nc, SubjectUserDelete, req)
	assertRPCSuccess(t, resp)
}

func TestRPC_DeleteUser_NotFound(t *testing.T) {
	q := &testutil.MockQuerier{
		UserExistsFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
	}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := DeleteUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     testutil.TestUserID.String(),
	}
	resp := rpcRequest(t, nc, SubjectUserDelete, req)
	assertRPCError(t, resp, http.StatusNotFound)
}

func TestRPC_DeleteUser_InvalidUUID(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	req := DeleteUserRPCRequest{
		RPCRequest: RPCRequest{ActorID: "actor-1"},
		UserID:     "bad-uuid",
	}
	resp := rpcRequest(t, nc, SubjectUserDelete, req)
	assertRPCError(t, resp, http.StatusBadRequest)
}

func TestRPC_DeleteUser_MalformedJSON(t *testing.T) {
	q := &testutil.MockQuerier{}
	_, nc, cleanup := setupRPCServer(t, q)
	defer cleanup()

	resp := sendRawAndDecode(t, nc, SubjectUserDelete, []byte("{bad}"))
	assertRPCError(t, resp, http.StatusBadRequest)
}
