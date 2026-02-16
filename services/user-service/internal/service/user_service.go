package service

import (
	"context"
	"errors"
	"log"

	database "user-service/db/sqlc"
	"user-service/internal/models"
	"user-service/internal/nats/publisher"
	"user-service/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	pool      *pgxpool.Pool
	queries   database.Querier
	publisher publisher.EventPublisherInterface
}

func NewUserService(pool *pgxpool.Pool, queries database.Querier, pub publisher.EventPublisherInterface) *UserService {
	return &UserService{
		pool:      pool,
		queries:   queries,
		publisher: pub,
	}
}

// CreateUser creates a new user and publishes a creation event
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest, actorID string) (*models.UserResponse, error) {
	exists, err := s.queries.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to check email existence", err)
	}
	if exists {
		return nil, models.NewConflictError("Email already exists")
	}

	status := req.Status
	if status == "" {
		status = models.UserStatusActive
	}

	// Build create parameters
	params := database.CreateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     utils.ConvertStringPtrToText(req.Phone),
		Age:       utils.ConvertIntPtrToInt4(req.Age),
		Status:    string(status),
	}

	user, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to create user", err)
	}

	userResponse := utils.ConvertToUserResponse(user)

	// Publish user created event
	if pubErr := s.publisher.PublishUserCreated(ctx, actorID, userResponse); pubErr != nil {
		// for now log error but don't fail the operation
		log.Printf("Failed to publish user created event: %v", pubErr)
	}

	return userResponse, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.UserResponse, error) {
	// Parse UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.NewBadRequestError("Invalid user ID format")
	}

	// Query database
	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.NewNotFoundError("User not found")
		}
		return nil, models.NewInternalServerError("Failed to get user", err)
	}

	return utils.ConvertToUserResponse(user), nil
}

// ListUsers retrieves all users
func (s *UserService) ListUsers(ctx context.Context) (*models.ListUsersResponse, error) {
	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to list users", err)
	}

	// Convert to response slice
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *utils.ConvertToUserResponse(user)
	}

	return &models.ListUsersResponse{
		Users: userResponses,
		Total: len(userResponses),
	}, nil
}

// UpdateUser updates an existing user and publishes an update event
func (s *UserService) UpdateUser(ctx context.Context, userID string, req models.UpdateUserRequest, actorID string) (*models.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.NewBadRequestError("Invalid user ID format")
	}

	exists, err := s.queries.UserExists(ctx, id)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to check user", err)
	}
	if !exists {
		return nil, models.NewNotFoundError("User not found")
	}

	if req.Email != nil {
		emailExists, err := s.queries.EmailExists(ctx, *req.Email)
		if err != nil {
			return nil, models.NewInternalServerError("Failed to check email", err)
		}

		currentUser, err := s.queries.GetUserByID(ctx, id)
		if err != nil {
			return nil, models.NewInternalServerError("Failed to get user", err)
		}

		if emailExists && currentUser.Email != *req.Email {
			return nil, models.NewConflictError("Email already exists")
		}
	}

	// Build update parameters
	params := database.UpdateUserParams{
		UserID:    id,
		FirstName: utils.ConvertStringPtrToText(req.FirstName),
		LastName:  utils.ConvertStringPtrToText(req.LastName),
		Email:     utils.ConvertStringPtrToText(req.Email),
		Phone:     utils.ConvertStringPtrToText(req.Phone),
		Age:       utils.ConvertIntPtrToInt4(req.Age),
		Status: func() database.NullUserStatus {
			if req.Status != nil {
				return database.NullUserStatus{
					UserStatus: database.UserStatus(*req.Status),
					Valid:      true,
				}
			}
			return database.NullUserStatus{Valid: false}
		}(),
	}

	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to update user", err)
	}

	userResponse := utils.ConvertToUserResponse(user)

	// Publish user updated event
	if pubErr := s.publisher.PublishUserUpdated(ctx, actorID, userResponse); pubErr != nil {
		log.Printf("Failed to publish user updated event: %v", pubErr)
	}

	return userResponse, nil
}

// DeleteUser deletes a user and publishes a deletion event
func (s *UserService) DeleteUser(ctx context.Context, userID string, actorID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return models.NewBadRequestError("Invalid user ID format")
	}

	exists, err := s.queries.UserExists(ctx, id)
	if err != nil {
		return models.NewInternalServerError("Failed to check user", err)
	}
	if !exists {
		return models.NewNotFoundError("User not found")
	}

	err = s.queries.DeleteUser(ctx, id)
	if err != nil {
		return models.NewInternalServerError("Failed to delete user", err)
	}

	// Publish user deleted event
	if pubErr := s.publisher.PublishUserDeleted(ctx, actorID, userID); pubErr != nil {
		log.Printf("Failed to publish user deleted event: %v", pubErr)
	}

	return nil
}
