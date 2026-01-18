package postgres

import (
	"context"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

type MessageSchema struct {
	gorm.Model
	ID                 string    `json:"id" gorm:"primaryKey"`
	SenderID           string    `json:"sender_id" gorm:"not null"`
	ReceiverID         string    `json:"receiver_id" gorm:"not null"`
	Content            string    `json:"content" gorm:"not null"`
	Type               string    `json:"type" gorm:"not null"`               // 'general' | 'invitation' | invitation_accepted' | 'invitation_rejected' |
	InvitationGrindID  string    `json:"invitation_grind_id" gorm:""`        // the id of the grind that the invitation is for
	InvitationAccepted bool      `json:"invitation_accepted" gorm:""`        // whether the invitation has been accepted by the receiver
	InvitationRejected bool      `json:"invitation_rejected" gorm:""`        // whether the invitation has been rejected by the receiver
	Read               bool      `json:"read" gorm:"not null;default:false"` // whether the message has been read by the receiver
	CreatedAt          time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"not null"`
}

func (MessageSchema) TableName() string { return "message" }

type GormMessageRepository struct {
	db *gorm.DB
}

func NewGormMessageRepository(db *gorm.DB) *GormMessageRepository {
	return &GormMessageRepository{db: db}
}

func (r *GormMessageRepository) Create(message *entities.Message) error {
	ctx := context.Background()
	model := MessageSchema{
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
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormMessageRepository) FindByID(id string) (*entities.Message, error) {
	ctx := context.Background()
	var model MessageSchema
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entities.Message{
		ID:                 model.ID,
		SenderID:           model.SenderID,
		ReceiverID:         model.ReceiverID,
		Content:            model.Content,
		Type:               model.Type,
		InvitationGrindID:  model.InvitationGrindID,
		InvitationAccepted: model.InvitationAccepted,
		InvitationRejected: model.InvitationRejected,
		Read:               model.Read,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}, nil
}

func (r *GormMessageRepository) FindAllForReceiver(receiverID string, offset, limit int) ([]*entities.Message, error) {
	ctx := context.Background()
	var models []MessageSchema
	query := r.db.WithContext(ctx).Where("receiver_id = ?", receiverID).Order("created_at DESC")
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&models).Error
	if err != nil {
		return nil, err
	}
	
	messages := make([]*entities.Message, len(models))
	for i, model := range models {
		messages[i] = &entities.Message{
			ID:                 model.ID,
			SenderID:           model.SenderID,
			ReceiverID:         model.ReceiverID,
			Content:            model.Content,
			Type:               model.Type,
			InvitationGrindID:  model.InvitationGrindID,
			InvitationAccepted: model.InvitationAccepted,
			InvitationRejected: model.InvitationRejected,
			Read:               model.Read,
			CreatedAt:          model.CreatedAt,
			UpdatedAt:          model.UpdatedAt,
		}
	}
	return messages, nil
}

func (r *GormMessageRepository) Update(message *entities.Message) error {
	ctx := context.Background()
	model := MessageSchema{
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
		UpdatedAt:          time.Now().UTC(),
	}
	return r.db.WithContext(ctx).Model(&MessageSchema{}).Where("id = ?", message.ID).Updates(&model).Error
}
