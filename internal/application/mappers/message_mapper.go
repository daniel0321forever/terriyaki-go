package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildMessageDTO constructs Message DTO from Message-related entities
func BuildMessageDTO(message *entities.Message, sender *entities.User, receiver *entities.User, invitationGrind *entities.Grind) *dto.MessageDTO {
	var grindDTO *dto.MessageGrindDTO
	if invitationGrind != nil {
		grindDTO = BuildMessageGrindDTO(invitationGrind)
	}

	return &dto.MessageDTO{
		ID:                 message.ID,
		Sender:             BuildUserDTO(sender),
		Receiver:           BuildUserDTO(receiver),
		InvitationGrind:    grindDTO,
		Content:            message.Content,
		Type:               message.Type,
		InvitationAccepted: message.InvitationAccepted,
		InvitationRejected: message.InvitationRejected,
		Read:               message.Read,
		CreatedAt:          message.CreatedAt,
		UpdatedAt:          message.UpdatedAt,
	}
}
