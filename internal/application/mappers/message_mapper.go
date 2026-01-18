package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// MessageToMessageDTO converts Message entity to MessageDTO
func MessageToMessageDTO(message *entities.Message) *dto.MessageDTO {
	return &dto.MessageDTO{
		ID:                 message.ID,
		SenderID:           message.SenderID,
		ReceiverID:         message.ReceiverID,
		Content:            message.Content,
		Type:               message.Type,
		InvitationGrindID:  message.InvitationGrindID,
		InvitationAccepted: message.InvitationAccepted,
		InvitationRejected: message.InvitationRejected,
		Read:               message.Read,
		CreatedAt:          message.CreatedAt,
		UpdatedAt:          message.UpdatedAt,
	}
}

