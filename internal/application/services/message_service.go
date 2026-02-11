package services

import (
	"errors"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
)

type MessageService struct {
	messageRepo repositories.MessageRepository
	userRepo    repositories.UserRepository
	grindRepo   repositories.GrindRepository
}

func NewMessageService(messageRepo repositories.MessageRepository, userRepo repositories.UserRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

func (s *MessageService) CreateInvitationMessage(request dto.CreateInvitationMessageDTO) (*dto.MessageDTO, error) {
	// get receiver
	receiver, err := s.userRepo.FindByEmail(request.ReceiverEmail)
	if err != nil {
		return nil, err
	}

	// Create message entity using constructor
	message, err := entities.NewMessage(
		request.SenderID,
		receiver.ID,
		request.SenderID+" invited you to join a grind",
		"invitation",
		"",
		false, // invitationAccepted
		false, // invitationRejected
	)
	if err != nil {
		return nil, err
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, err
	}

	return mappers.MessageToMessageDTO(message), nil
}

func (s *MessageService) CreateInvitationAcceptedMessage(request dto.CreateInvitationAcceptedMessageDTO) (*dto.MessageDTO, error) {

	// Create message entity using constructor
	message, err := entities.NewMessage(
		request.AccepterID,
		request.InvitorID,
		request.AccepterID+" accepted your invitation",
		"invitation_accepted",
		request.GrindID,
		true,  // invitationAccepted
		false, // invitationRejected
	)
	if err != nil {
		return nil, err
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, err
	}

	return mappers.MessageToMessageDTO(message), nil
}

func (s *MessageService) CreateInvitationRejectedMessage(request dto.CreateInvitationRejectedMessageDTO) (*dto.MessageDTO, error) {
	// Create message entity using constructor
	message, err := entities.NewMessage(
		request.RejecterID,
		request.InvitorID,
		request.RejecterID+" rejected your invitation",
		"invitation_rejected",
		request.GrindID,
		false, // invitationAccepted
		true,  // invitationRejected
	)
	if err != nil {
		return nil, err
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, err
	}

	return mappers.MessageToMessageDTO(message), nil
}

func (s *MessageService) GetMessageByID(request dto.GetMessageDTO) (*dto.MessageDTO, error) {
	message, err := s.messageRepo.FindByID(request.MessageID)
	if err != nil {
		return nil, errors.New("grind not found")
	}
	return mappers.MessageToMessageDTO(message), nil
}

func (s *MessageService) GetAllMessagesForReceiver(request dto.GetAllMessagesForReceiverDTO) ([]*dto.MessageDTO, error) {
	messages, err := s.messageRepo.FindAllForReceiver(request.ReceiverID, request.Offset, request.Limit)
	if err != nil {
		return nil, errors.New("message not found")
	}
	var output []*dto.MessageDTO
	for _, message := range messages {
		output = append(output, mappers.MessageToMessageDTO(message))
	}
	return output, nil
}

func (s *MessageService) UpdateMessageReadStatus(request dto.UpdateMessageReadStatusDTO) (*dto.MessageDTO, error) {
	message, err := s.messageRepo.FindByID(request.MessageID)
	if err != nil {
		return nil, err
	}

	message.Read = request.Read
	message.UpdatedAt = time.Now().UTC()

	err = s.messageRepo.Update(message)
	if err != nil {
		return nil, err
	}

	return mappers.MessageToMessageDTO(message), nil
}

func (s *MessageService) UpdateMessageInvitationAcceptedStatus(request dto.UpdateMessageInvitationAcceptedStatusDTO) (*dto.MessageDTO, error) {
	message, err := s.messageRepo.FindByID(request.MessageID)
	if err != nil {
		return nil, err
	}

	message.InvitationAccepted = request.Accepted
	message.UpdatedAt = time.Now().UTC()

	err = s.messageRepo.Update(message)
	if err != nil {
		return nil, err
	}

	return mappers.MessageToMessageDTO(message), nil
}

func (s *MessageService) GetAllMessageFromSender(senderID string, offset, limit int) ([]*dto.MessageDTO, error) {
	messages, err := s.messageRepo.FindAllFromSender(senderID, offset, limit)
	if err != nil {
		return nil, errors.New("message not found")
	}
	var output []*dto.MessageDTO
	for _, message := range messages {
		output = append(output, mappers.MessageToMessageDTO(message))
	}
	return output, nil
}
