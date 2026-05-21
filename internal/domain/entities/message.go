package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Message struct {
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

/** Constructor in factory pattern
 * @param senderID - the ID of the message sender
 * @param receiverID - the ID of the message receiver
 * @param content - the message content
 * @param messageType - the type of message: "general", "invitation", "invitation_accepted", "invitation_rejected"
 * @param invitationGrindID - optional: the grind ID for invitation-related messages
 * @param invitationAccepted - optional: whether invitation is accepted (for invitation_accepted type)
 * @param invitationRejected - optional: whether invitation is rejected (for invitation_rejected type)
 * @return the created message
 */
func NewMessage(
	senderID, receiverID, content, messageType string,
	invitationGrindID string,
	invitationAccepted, invitationRejected bool,
) (*Message, error) {
	// Validation
	if strings.TrimSpace(senderID) == "" {
		return nil, errors.New("senderID cannot be empty")
	}
	if strings.TrimSpace(receiverID) == "" {
		return nil, errors.New("receiverID cannot be empty")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content cannot be empty")
	}
	if strings.TrimSpace(messageType) == "" {
		return nil, errors.New("messageType cannot be empty")
	}

	// Validate message type
	validTypes := map[string]bool{
		"general":             true,
		"invitation":          true,
		"invitation_accepted": true,
		"invitation_rejected": true,
	}
	if !validTypes[messageType] {
		return nil, errors.New("invalid message type: must be 'general', 'invitation', 'invitation_accepted', or 'invitation_rejected'")
	}

	// Validate invitation-related fields based on type
	if messageType == "invitation" || messageType == "invitation_accepted" || messageType == "invitation_rejected" {
		if strings.TrimSpace(invitationGrindID) == "" {
			return nil, errors.New("invitationGrindID is required for invitation-related messages")
		}
	}

	now := time.Now().UTC()

	return &Message{
		ID:                 uuid.New().String(),
		SenderID:           strings.TrimSpace(senderID),
		ReceiverID:         strings.TrimSpace(receiverID),
		Content:            strings.TrimSpace(content),
		Type:               messageType,
		InvitationGrindID:  strings.TrimSpace(invitationGrindID),
		InvitationAccepted: invitationAccepted,
		InvitationRejected: invitationRejected,
		Read:               false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}
