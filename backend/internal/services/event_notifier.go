package services

import (
	"time"

	"github.com/alpyxn/aeterna/backend/internal/ports"
)

type eventNotifier struct {
	stream ports.EventStreamPort
}

func newEventNotifier(stream ports.EventStreamPort) eventNotifier {
	return eventNotifier{stream: stream}
}

func (n eventNotifier) publish(userID, eventType, resource, entityID, reason string) {
	if n.stream == nil || userID == "" || eventType == "" {
		return
	}
	n.stream.Publish(userID, ports.RealtimeEvent{
		Type:     eventType,
		At:       time.Now().UTC(),
		Resource: resource,
		EntityID: entityID,
		Reason:   reason,
	})
}
