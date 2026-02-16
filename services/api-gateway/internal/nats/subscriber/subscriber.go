package subscriber

import (
	"encoding/json"
	"log"
	"time"

	ws "api-gateway/internal/websocket"

	"github.com/nats-io/nats.go"
)

/* API GATEWAY SUBSCRIBER */

// NATS PubSub topics
const (
	TopicUserCreated = "user.events.created"
	TopicUserUpdated = "user.events.updated"
	TopicUserDeleted = "user.events.deleted"
)

type Event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	ActorID   string          `json:"actor_id"`
	UserID    string          `json:"user_id"`
	Data      json.RawMessage `json:"data"`
}

// * Subscriber listens to NATS PubSub events and broadcasts them via WebSocket
type Subscriber struct {
	nc      *nats.Conn           // nats connection
	manager *ws.Manager          // websocket manager to broadcast messages to clients
	subs    []*nats.Subscription // keep track of subscriptions for cleanup
}

// NewSubscriber creates a new NATS event subscriber
func NewSubscriber(nc *nats.Conn, manager *ws.Manager) *Subscriber {
	return &Subscriber{
		nc:      nc,
		manager: manager,
		subs:    make([]*nats.Subscription, 0),
	}
}

// Start subscribes to all user event topics
func (s *Subscriber) Start() error {

	// mapping each topic to its handler function - registering all handlers
	handlers := map[string]nats.MsgHandler{
		TopicUserCreated: s.handleUserCreated,
		TopicUserUpdated: s.handleUserUpdated,
		TopicUserDeleted: s.handleUserDeleted,
	}

	for topic, handler := range handlers {
		sub, err := s.nc.Subscribe(topic, handler)
		if err != nil {
			return err
		}

		// storing the subscription for later cleanup
		s.subs = append(s.subs, sub)
		log.Printf("Subscribed to event topic: %s", topic)
	}

	return nil
}

// Stop unsubscribes from all topics
func (s *Subscriber) Stop() error {
	for _, sub := range s.subs {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing from %s: %v", sub.Subject, err)
		}
	}
	return nil
}

/* EVENT HANDLERS */

// handleUserCreated broadcasts user.created events to ALL connected clients
func (s *Subscriber) handleUserCreated(msg *nats.Msg) {
	var event Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal user.created event: %v", err)
		return
	}

	log.Printf("Received user.created event (event_id: %s, user_id: %s)", event.EventID, event.UserID)

	wsMsg := &ws.WSMessage{
		Type:      "user.created",
		Timestamp: event.Timestamp,
		UserID:    event.UserID,
		Data:      json.RawMessage(event.Data),
	}

	// Broadcast to ALL connected clients
	s.manager.BroadcastToAll(wsMsg)
}

// handleUserUpdated broadcasts user.updated events to all clients
func (s *Subscriber) handleUserUpdated(msg *nats.Msg) {
	var event Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal user.updated event: %v", err)
		return
	}

	log.Printf("Received user.updated event (event_id: %s, user_id: %s, actor: %s)", event.EventID, event.UserID, event.ActorID)

	// transform the event into a WebSocket message format (Event -> WSMessage)
	wsMsg := &ws.WSMessage{
		Type:      "user.updated",
		Timestamp: event.Timestamp,
		UserID:    event.UserID,
		Data:      json.RawMessage(event.Data),
	}

	// Broadcast to all clients
	s.manager.BroadcastToAll(wsMsg)
}

// handleUserDeleted broadcasts user.deleted events to ALL connected clients except the user who was deleted
func (s *Subscriber) handleUserDeleted(msg *nats.Msg) {
	var event Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal user.deleted event: %v", err)
		return
	}

	log.Printf("Received user.deleted event (event_id: %s, user_id: %s)", event.EventID, event.UserID)

	wsMsg := &ws.WSMessage{
		Type:      "user.deleted",
		Timestamp: event.Timestamp,
		UserID:    event.UserID,
		Data:      json.RawMessage(event.Data),
	}

	// Broadcast to ALL connected clients except the user who was deleted
	s.manager.BroadcastExcluding(wsMsg, event.UserID)
}
