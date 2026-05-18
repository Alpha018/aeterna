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

func (n eventNotifier) publish(userID, eventType, eventCode, resource, entityID, reason string) {
	if n.stream == nil || userID == "" || eventType == "" {
		return
	}
	data := map[string]string{}
	if resource != "" {
		data["resource"] = resource
	}
	if entityID != "" {
		data["entity_id"] = entityID
	}
	if reason != "" {
		data["reason"] = reason
	}
	if len(data) == 0 {
		data = nil
	}

	n.stream.Publish(userID, ports.RealtimeEvent{
		Type:     eventType,
		Code:     eventCode,
		At:       time.Now().UTC(),
		Data:     data,
		Resource: resource,
		EntityID: entityID,
		Reason:   reason,
	})
}
