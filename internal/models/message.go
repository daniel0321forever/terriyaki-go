package models

import (
	"errors"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
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

/**
 * Create a new message
 * @param senderID - the id of the sender
 * @param receiverID - the id of the receiver
 * @param content - the content of the message
 * @param type - the type of the message
 * @param invitationGrindID - the id of the grind that the invitation is for
 * @return the created message
 */
func CreateMessage(senderID string, receiverID string, content string, messageType string, invitationGrindID string) (*Message, error) {
	message := Message{
		ID:                 uuid.New().String(),
		SenderID:           senderID,
		ReceiverID:         receiverID,
		Content:            content,
		Type:               messageType,
		InvitationGrindID:  invitationGrindID,
		InvitationAccepted: false,
		InvitationRejected: false,
		Read:               false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	result := database.Db.Create(&message)
	if result.Error != nil {
		return nil, result.Error
	}

	return &message, nil
}

/**
 * Create an invitation message
 * @param senderID - the id of the sender
 * @param receiverID - the id of the receiver
 * @param invitationGrindID - the id of the grind that the invitation is for
 * @return the created message
 */
func CreateInvitationMessage(sender *User, receiver *User, invitationGrindID string) (*Message, error) {
	if sender.ID == receiver.ID {
		return nil, errors.New(config.ERROR_CODE_SAME_RECIPIENT_AND_SENDER)
	}

	message := Message{
		ID:                uuid.New().String(),
		SenderID:          sender.ID,
		ReceiverID:        receiver.ID,
		Content:           "You have been invited to a grind created by " + sender.Username,
		Type:              config.MESSAGE_TYPE_INVITATION,
		InvitationGrindID: invitationGrindID,
		Read:              false,
	}

	result := database.Db.Create(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	return &message, nil
}

/**
 * Create an invitation accepted message
 * @param senderID - the id of the sender
 * @param receiverID - the id of the receiver
 * @param invitationGrindID - the id of the grind that the invitation is for
 * @return the created message
 */
func CreateInvitationAcceptedMessage(sender *User, receiver *User, invitationGrindID string) (*Message, error) {
	if sender.ID == receiver.ID {
		return nil, errors.New(config.ERROR_CODE_SAME_RECIPIENT_AND_SENDER)
	}

	message := Message{
		ID:                uuid.New().String(),
		SenderID:          sender.ID,
		ReceiverID:        receiver.ID,
		Content:           "Your invitation to the grind has been accepted by " + sender.Username,
		Type:              config.MESSAGE_TYPE_INVITATION_ACCEPTED,
		InvitationGrindID: invitationGrindID,
		Read:              false,
	}

	result := database.Db.Create(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	return &message, nil
}

/**
 * Create an invitation rejected message
 * @param senderID - the id of the sender
 * @param receiverID - the id of the receiver
 * @param invitationGrindID - the id of the grind that the invitation is for
 * @return the created message
 */
func CreateInvitationRejectedMessage(sender *User, receiver *User, invitationGrindID string) (*Message, error) {
	if sender.ID == receiver.ID {
		return nil, errors.New(config.ERROR_CODE_SAME_RECIPIENT_AND_SENDER)
	}

	message := Message{
		ID:                uuid.New().String(),
		SenderID:          sender.ID,
		ReceiverID:        receiver.ID,
		Content:           "Your invitation to the grind has been rejected by " + sender.Username,
		Type:              config.MESSAGE_TYPE_INVITATION_REJECTED,
		InvitationGrindID: invitationGrindID,
		Read:              false,
	}

	result := database.Db.Create(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	return &message, nil
}

/**
 * Get a message by id
 * @param id - the id of the message
 * @return the message
 */
func GetMessageByID(id string) (*Message, error) {
	var message Message
	result := database.Db.Where("id = ?", id).First(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	return &message, nil
}

/**
 * Get all messages for a receiver
 * @param receiverID - the id of the receiver
 * @return the messages
 */
func GetAllMessageForReceiver(receiverID string, offset int, limit int) ([]Message, error) {
	var messages []Message
	result := database.Db.Where("receiver_id = ?", receiverID).Offset(offset).Limit(limit).Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}
	return messages, nil
}

/**
 * Get all messages from a sender
 * @param senderID - the id of the sender
 * @return the messages
 */
func GetAllMessageFromSender(senderID string, offset int, limit int) ([]Message, error) {
	var messages []Message
	result := database.Db.Where("sender_id = ?", senderID).Offset(offset).Limit(limit).Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}
	return messages, nil
}

/**
 * Get an invitation message
 * @param invitorID - the id of the invitor
 * @param participantID - the id of the participant
 * @param invitationGrindID - the id of the grind that the invitation is for
 * @return the invitation message
 */
func GetInvitationMessage(invitorID string, participantID string, invitationGrindID string) (*Message, error) {
	var messages []Message
	result := database.Db.Where("sender_id = ? AND receiver_id = ? AND invitation_grind_id = ?", invitorID, participantID, invitationGrindID).Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}
	if len(messages) == 0 {
		return nil, errors.New(config.ERROR_CODE_INVITATION_MESSAGE_NOT_FOUND)
	}
	return &messages[0], nil
}

/**
 * Update a message
 * @param id - the id of the message
 * @param read - whether the message has been read by the receiver
 * @return the updated message
 */
func UpdateMessageReadStatus(id string, read bool) (*Message, error) {
	var message Message
	result := database.Db.Where("id = ?", id).First(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	message.Read = read
	result = database.Db.Save(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	return &message, nil
}

/**
 * Update a message invitation responded status
 * @param id - the id of the message
 * @param responded - whether the invitation has been responded by the receiver
 * @return the updated message
 */
func UpdateMessageInvitationAcceptedStatus(id string, accepted bool) (*Message, error) {
	message := &Message{}
	result := database.Db.Model(&message).Where("id = ?", id).Update("invitation_accepted", accepted).Update("invitation_rejected", !accepted)
	if result.Error != nil {
		return nil, result.Error
	}
	return message, nil
}
