package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/container"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// MessageToMessageDTO converts Message entity to MessageDTO
func MessageToMessageDTO(message *entities.Message) *dto.MessageDTO {
	// TODO: it is actually kinda bad to repeatedly fetch data when getting message, we should optimize it in repository to avoid multiple database queries
	sender, err := container.Repos.UserRepository.FindById(message.SenderID)
	if err != nil {
		panic(err)
	}
	receiver, err := container.Repos.UserRepository.FindById(message.ReceiverID)
	if err != nil {
		panic(err)
	}

	var grindDTO *dto.MessageGrindDTO = nil
	if message.InvitationGrindID != "" {
		grind, err := container.Repos.GrindRepository.FindById(message.InvitationGrindID)
		if err == nil {
			grindDTO = GrindToMessageGrindDTO(grind)
		}
	}

	return &dto.MessageDTO{
		ID:                 message.ID,
		Sender:             UserToUserDTO(sender),
		Receiver:           UserToUserDTO(receiver),
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
