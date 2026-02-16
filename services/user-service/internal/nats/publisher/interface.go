package publisher

import (
	"context"
	"user-service/internal/models"
)

// contract for publishing domain events, allows the publisher to be mocked in unit tests.
type EventPublisherInterface interface {
	PublishUserCreated(ctx context.Context, actorID string, user *models.UserResponse) error
	PublishUserUpdated(ctx context.Context, actorID string, user *models.UserResponse) error
	PublishUserDeleted(ctx context.Context, actorID string, userID string) error
}
