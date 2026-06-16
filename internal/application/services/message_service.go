package services

import (
	"errors"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"gorm.io/gorm"
)

type MessageService struct {
	db          *gorm.DB
	messageRepo repositories.MessageRepository
	userRepo    repositories.UserRepository
	grindRepo   repositories.GrindRepository
}

func NewMessageService(db *gorm.DB, messageRepo repositories.MessageRepository, userRepo repositories.UserRepository, grindRepo repositories.GrindRepository) *MessageService {
	return &MessageService{
		db:          db,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		grindRepo:   grindRepo,
	}
}

// Convert Message entity to Message DTO (including related entity fetching from DB)
func (s *MessageService) toMessageDTO(message *entities.Message) (*dto.MessageDTO, error) {
	sender, err := s.userRepo.FindById(message.SenderID)
	if err != nil {
		return nil, err
	}

	receiver, err := s.userRepo.FindById(message.ReceiverID)
	if err != nil {
		return nil, err
	}

	var invitationGrind *entities.Grind
	if message.InvitationGrindID != "" {
		grind, err := s.grindRepo.FindById(message.InvitationGrindID)
		if err == nil {
			invitationGrind = grind
		}
	}

	return mappers.BuildMessageDTO(message, sender, receiver, invitationGrind), nil
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

	return s.toMessageDTO(message)
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

	return s.toMessageDTO(message)
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

	return s.toMessageDTO(message)
}

func (s *MessageService) GetMessageByID(request dto.GetMessageDTO) (*dto.MessageDTO, error) {
	message, err := s.messageRepo.FindByID(request.MessageID)
	if err != nil {
		return nil, errors.New("grind not found")
	}
	return s.toMessageDTO(message)
}

func (s *MessageService) GetAllMessagesForReceiver(request dto.GetAllMessagesForReceiverDTO) ([]*dto.MessageDTO, error) {
	messages, err := s.messageRepo.FindAllForReceiver(request.ReceiverID, request.Offset, request.Limit)
	if err != nil {
		return nil, errors.New("message not found")
	}
	var output []*dto.MessageDTO
	for _, message := range messages {
		messageDTO, dtoErr := s.toMessageDTO(message)
		if dtoErr != nil {
			return nil, dtoErr
		}
		output = append(output, messageDTO)
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

	return s.toMessageDTO(message)
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

	return s.toMessageDTO(message)
}

func (s *MessageService) GetAllMessageFromSender(senderID string, offset, limit int) ([]*dto.MessageDTO, error) {
	messages, err := s.messageRepo.FindAllFromSender(senderID, offset, limit)
	if err != nil {
		return nil, errors.New("message not found")
	}
	var output []*dto.MessageDTO
	for _, message := range messages {
		messageDTO, dtoErr := s.toMessageDTO(message)
		if dtoErr != nil {
			return nil, dtoErr
		}
		output = append(output, messageDTO)
	}
	return output, nil
}

// RejectInvitationTx executes UpdateMessageInvitationAcceptedStatus and CreateInvitationRejectedMessage
// atomically in a single DB transaction.
func (s *MessageService) RejectInvitationTx(
	updateReq dto.UpdateMessageInvitationAcceptedStatusDTO,
	createReq dto.CreateInvitationRejectedMessageDTO,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		msgRepo := getMessageRepo(s.messageRepo, tx)

		// Update original invitation message to rejected
		msg, err := msgRepo.FindByID(updateReq.MessageID)
		if err != nil {
			return err
		}
		msg.InvitationAccepted = false
		msg.InvitationRejected = true
		msg.UpdatedAt = time.Now().UTC()
		if err := msgRepo.Update(msg); err != nil {
			return err
		}

		// Create rejection notification message to invitor
		rejectedMsg, err := entities.NewMessage(
			createReq.RejecterID,
			createReq.InvitorID,
			createReq.RejecterID+" rejected your invitation",
			"invitation_rejected",
			createReq.GrindID,
			false,
			true,
		)
		if err != nil {
			return err
		}
		return msgRepo.Create(rejectedMsg)
	})
}
