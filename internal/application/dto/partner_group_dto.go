package dto

import "time"

// CreatePartnerGroupDTO is the request body for creating a partner group.
type CreatePartnerGroupDTO struct {
	GrindID string `json:"grindID"`
	Name    string `json:"name"`
}

// PartnerGroupDTO is the response DTO for a PartnerGroup entity.
type PartnerGroupDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	GrindID   string    `json:"grindID"`
	OwnerID   string    `json:"ownerID"`
	Members   []string  `json:"members"`
	CreatedAt time.Time `json:"createdAt"`
}

// InviteLinkDTO is the response DTO for a generated invite link.
type InviteLinkDTO struct {
	Token     string `json:"token"`
	InviteURL string `json:"inviteURL"`
}

// JoinGroupDTO is the request body for joining a partner group.
type JoinGroupDTO struct {
	Token string `json:"token"`
}
