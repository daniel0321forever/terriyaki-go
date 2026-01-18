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
}

func NewMessageService(messageRepo repositories.MessageRepository, userRepo repositories.UserRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

func (s *MessageService) CreateInvitationMessage(request dto.CreateInvitationMessageDTO) (*dto.MessageDTO, error) {
	sender, err := s.userRepo.FindById(request.SenderID)
	if err != nil {
		return nil, err
	}

	// Create message entity using constructor
	message, err := entities.NewMessage(
		request.SenderID,
		request.ReceiverID,
		sender.Username+" invited you to join a grind",
		"invitation",
		request.GrindID,
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
	accepter, err := s.userRepo.FindById(request.AccepterID)
	if err != nil {
		return nil, err
	}

	// Create message entity using constructor
	message, err := entities.NewMessage(
		request.AccepterID,
		request.InvitorID,
		accepter.Username+" accepted your invitation",
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
	rejecter, err := s.userRepo.FindById(request.RejecterID)
	if err != nil {
		return nil, err
	}

	// Create message entity using constructor
	message, err := entities.NewMessage(
		request.RejecterID,
		request.InvitorID,
		rejecter.Username+" rejected your invitation",
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

