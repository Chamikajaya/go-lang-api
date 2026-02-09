package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"user-service/internal/models"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// NATS PubSub topics for events
const (
	TopicUserCreated = "user.events.created"
	TopicUserUpdated = "user.events.updated"
	TopicUserDeleted = "user.events.deleted"
)

// EventPublisher publishes domain events to NATS
type EventPublisher struct {
	nc *nats.Conn
}

// NewEventPublisher creates a new EventPublisher
func NewEventPublisher(nc *nats.Conn) *EventPublisher {
	return &EventPublisher{
		nc: nc,
	}
}

// PublishUserCreated publishes a user created event
func (p *EventPublisher) PublishUserCreated(ctx context.Context, actorID string, user *models.UserResponse) error {
	event := &models.Event{
		EventID:   uuid.New(),
		EventType: models.EventTypeUserCreated,
		Timestamp: time.Now().UTC(),
		ActorID:   actorID,
		UserID:    user.ID.String(),
		Data: models.UserCreatedEventData{
			User: *user,
		},
	}

	return p.publish(TopicUserCreated, event)
}

// PublishUserUpdated publishes a user updated event
func (p *EventPublisher) PublishUserUpdated(ctx context.Context, actorID string, user *models.UserResponse) error {
	event := &models.Event{
		EventID:   uuid.New(),
		EventType: models.EventTypeUserUpdated,
		Timestamp: time.Now().UTC(),
		ActorID:   actorID,
		UserID:    user.ID.String(),
		Data: models.UserUpdatedEventData{
			User: *user,
		},
	}

	return p.publish(TopicUserUpdated, event)
}

// PublishUserDeleted publishes a user deleted event
func (p *EventPublisher) PublishUserDeleted(ctx context.Context, actorID string, userID string) error {
	event := &models.Event{
		EventID:   uuid.New(),
		EventType: models.EventTypeUserDeleted,
		Timestamp: time.Now().UTC(),
		ActorID:   actorID,
		UserID:    userID,
		Data: models.UserDeletedEventData{
			UserID: userID,
		},
	}

	return p.publish(TopicUserDeleted, event)
}

// publish serializes and publishes an event to NATS
func (p *EventPublisher) publish(topic string, event *models.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.nc.Publish(topic, data); err != nil {
		return fmt.Errorf("failed to publish event to %s: %w", topic, err)
	}

	log.Printf("Published event %s to topic %s (event_id: %s)", event.EventType, topic, event.EventID)
	return nil
}
