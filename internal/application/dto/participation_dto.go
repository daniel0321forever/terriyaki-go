package dto

import "time"

// Input DTOs
type GetParticipation struct {
	ParticipationID string
}

type GetParticipationByUserAndGrindDTO struct {
	UserID  string
	GrindID string
}

type AddParticipationDTO struct {
	GrindID string
	UserID  string
}

type UpdateAddParticipationDTO struct {
	ParticipationID string
	MissedDays      int
	TotalPenalty    int
	Quitted         bool
	QuittedAt       *time.Time
}

// Output DTOs
type ParticipationDTO struct {
	ID           string     `json:"id,omitempty"`
	UserID       string     `json:"userID"`
	GrindID      string     `json:"grindID"`
	MissedDays   int        `json:"missedDays"`
	TotalPenalty int        `json:"totalPenalty"`
	Quitted      bool       `json:"quitted"`
	QuittedAt    *time.Time `json:"quittedAt,omitempty"`
}
