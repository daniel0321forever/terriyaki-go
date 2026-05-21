package dto

import "time"

// Input DTOs
type CreateInvitationMessageDTO struct {
	SenderID      string
	ReceiverEmail string
	GrindID       string
}

type CreateInvitationAcceptedMessageDTO struct {
	AccepterID string
	InvitorID  string
	GrindID    string
}

type CreateInvitationRejectedMessageDTO struct {
	RejecterID string
	InvitorID  string
	GrindID    string
}

type GetMessageDTO struct {
	MessageID string
}

type GetAllMessagesForReceiverDTO struct {
	ReceiverID string
	Offset     int
	Limit      int
}

type UpdateMessageReadStatusDTO struct {
	MessageID string
	Read      bool
}

type UpdateMessageInvitationAcceptedStatusDTO struct {
	MessageID string
	Accepted  bool
}

// Output DTOs
type MessageDTO struct {
	ID                 string           `json:"id"`
	Sender             *UserDTO         `json:"sender"`
	Receiver           *UserDTO         `json:"receiver"`
	Content            string           `json:"content"`
	Type               string           `json:"type"` // "general", "invitation", "invitation_accepted", "invitation_rejected"
	InvitationGrind    *MessageGrindDTO `json:"grind,omitempty"`
	InvitationAccepted bool             `json:"invitationAccepted"`
	InvitationRejected bool             `json:"invitationRejected"`
	Read               bool             `json:"read"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
}
