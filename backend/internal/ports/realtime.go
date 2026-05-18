package ports

import "time"

const (
	EventTypeReady              = "ready"
	EventTypePing               = "ping"
	EventTypeMessagesChanged    = "messages.changed"
	EventTypeAttachmentsChanged = "attachments.changed"
	EventTypeFarewellsChanged   = "farewells.changed"
	EventTypeSettingsChanged    = "settings.changed"
	EventTypeWebhooksChanged    = "webhooks.changed"
)

// RealtimeEvent is delivered to authenticated SSE clients.
type RealtimeEvent struct {
	Type     string    `json:"type"`
	At       time.Time `json:"at"`
	Resource string    `json:"resource,omitempty"`
	EntityID string    `json:"entity_id,omitempty"`
	Reason   string    `json:"reason,omitempty"`
}

// EventStreamPort exposes user-scoped pub/sub for real-time refresh hints.
type EventStreamPort interface {
	Subscribe(userID, clientID string) (<-chan RealtimeEvent, <-chan struct{}, func(), error)
	Publish(userID string, event RealtimeEvent)
}
