package models

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeUserCreated EventType = "user.created"
	EventTypeUserUpdated EventType = "user.updated"
	EventTypeUserDeleted EventType = "user.deleted"
)

// Event represents a domain event published to NATS
type Event struct {
	EventID   uuid.UUID   `json:"event_id"`
	EventType EventType   `json:"event_type"`
	Timestamp time.Time   `json:"timestamp"`
	ActorID   string      `json:"actor_id"` // ID of user who triggered the event
	UserID    string      `json:"user_id"`  // ID of the affected user
	Data      interface{} `json:"data"`     // Event-specific data
}

// NewEvent creates a new event with auto-generated ID and timestamp
func NewEvent(eventType EventType, actorID, userID string, data interface{}) *Event {
	return &Event{
		EventID:   uuid.New(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		ActorID:   actorID,
		UserID:    userID,
		Data:      data,
	}
}

// UserCreatedEventData contains data for user created events
type UserCreatedEventData struct {
	User UserResponse `json:"user"`
}

// UserUpdatedEventData contains data for user updated events
type UserUpdatedEventData struct {
	User UserResponse `json:"user"`
}

// UserDeletedEventData contains data for user deleted events
type UserDeletedEventData struct {
	UserID string `json:"user_id"`
}
