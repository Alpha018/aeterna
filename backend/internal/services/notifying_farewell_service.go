package services

import (
	"github.com/alpyxn/aeterna/backend/internal/models"
	"github.com/alpyxn/aeterna/backend/internal/ports"
)

type NotifyingFarewellService struct {
	base     ports.FarewellServicePort
	notifier eventNotifier
}

func NewNotifyingFarewellService(base ports.FarewellServicePort, stream ports.EventStreamPort) ports.FarewellServicePort {
	return &NotifyingFarewellService{base: base, notifier: newEventNotifier(stream)}
}

func (s *NotifyingFarewellService) Create(userID, messageID, recipientEmail, subject, content string, delayMinutes int) (models.FarewellLetter, error) {
	letter, err := s.base.Create(userID, messageID, recipientEmail, subject, content, delayMinutes)
	if err == nil {
		s.notifier.publish(userID, ports.EventTypeFarewellsChanged, "farewell", letter.ID, "created")
		s.notifier.publish(userID, ports.EventTypeMessagesChanged, "message", messageID, "farewell_created")
	}
	return letter, err
}

func (s *NotifyingFarewellService) List(userID, messageID string) ([]models.FarewellLetter, error) {
	return s.base.List(userID, messageID)
}

func (s *NotifyingFarewellService) Update(userID, messageID, id, recipientEmail, subject, content string, delayMinutes int) (models.FarewellLetter, error) {
	letter, err := s.base.Update(userID, messageID, id, recipientEmail, subject, content, delayMinutes)
	if err == nil {
		s.notifier.publish(userID, ports.EventTypeFarewellsChanged, "farewell", letter.ID, "updated")
		s.notifier.publish(userID, ports.EventTypeMessagesChanged, "message", messageID, "farewell_updated")
	}
	return letter, err
}

func (s *NotifyingFarewellService) Delete(userID, messageID, id string) error {
	err := s.base.Delete(userID, messageID, id)
	if err == nil {
		s.notifier.publish(userID, ports.EventTypeFarewellsChanged, "farewell", id, "deleted")
		s.notifier.publish(userID, ports.EventTypeMessagesChanged, "message", messageID, "farewell_deleted")
	}
	return err
}
