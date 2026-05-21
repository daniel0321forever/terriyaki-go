package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// The interface: should be implemented by infrastructure/db/*
type MessageRepository interface {
	Create(message *entities.Message) error
	FindByID(id string) (*entities.Message, error)
	FindAllForReceiver(receiverID string, offset, limit int) ([]*entities.Message, error)
	FindAllFromSender(senderID string, offset, limit int) ([]*entities.Message, error)
	Update(message *entities.Message) error
}
