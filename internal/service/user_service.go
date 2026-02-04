package service

import (
	"context"
	"errors"

	database "user-management-api/db/sqlc"
	"user-management-api/internal/models"
	"user-management-api/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	pool    *pgxpool.Pool
	queries *database.Queries
}

// creating the user service instance - dependency injection
func NewUserService(pool *pgxpool.Pool, queries *database.Queries) *UserService {
	return &UserService{
		pool:    pool,
		queries: queries,
	}
}


func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {

	exists, err := s.queries.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to check email existence", err)
	}
	if exists {
		return nil, models.NewConflictError("Email Already Exists")
	}

	status := req.Status

	if status == "" {
		status = models.UserStatusActive
	}

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

	// Convert database model to response model
	return utils.ConvertToUserResponse(user), nil

}


func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.UserResponse, error) {
	
	// Parse UUID string to UUID type
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.NewBadRequestError("Invalid user ID format")
	}

	// Query database
	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		// pgx.ErrNoRows means not found
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.NewNotFoundError("User not found")
		}
		return nil, models.NewInternalServerError("Failed to get user", err)
	}

	return utils.ConvertToUserResponse(user), nil
}

// TODO: Add pagination later
func (s *UserService) ListUsers(ctx context.Context) (*models.ListUsersResponse, error) {
	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to list users", err)
	}

	// Convert slice of database users to response users
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *utils.ConvertToUserResponse(user)
	}

	return &models.ListUsersResponse{
		Users: userResponses,
		Total: len(userResponses),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, req models.UpdateUserRequest) (*models.UserResponse, error) {
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

	// If email is being updated, check for conflicts
	if req.Email != nil {
		emailExists, err := s.queries.EmailExists(ctx, *req.Email)
		if err != nil {
			return nil, models.NewInternalServerError("Failed to check email", err)
		}
		
		// Get current user to compare emails
		currentUser, err := s.queries.GetUserByID(ctx, id)
		if err != nil {
			return nil, models.NewInternalServerError("Failed to get user", err)
		}
		
		// Email exists and belongs to different user
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

	// Update in database
	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, models.NewInternalServerError("Failed to update user", err)
	}

	return utils.ConvertToUserResponse(user), nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
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

	return nil
}
