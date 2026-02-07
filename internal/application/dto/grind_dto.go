package dto

import (
	"time"
)

// Input DTOs
type CreateGrindDTO struct {
	CreatorID string    `json:"creator_id" validate:"required"`
	Duration  int       `json:"duration" validate:"min=1"`
	Budget    int       `json:"budget" validate:"min=0"`
	StartDate time.Time `json:"start_date"`
}

type GetGrindDTO struct {
	GrindID string
}

type DeleteGrindDTO struct {
	GrindID string
}

type GetOngoingGrindDTO struct {
	UserID string
}

type GetAllUserGrindsDTO struct {
	UserID string
}

type UpdateGrindDTO struct {
	// We usually get the ID from the URL path, not the body,
	// but keeping it here for batch updates is fine.
	GrindID  string `json:"-"`
	Duration int    `json:"duration,omitempty"`
	Budget   int    `json:"budget,omitempty"`
}

type QuitGrindDTO struct {
	UserID  string
	GrindID string
}

// Output DTOs
type GroupGrindDTO struct {
	ID        string    `json:"id"`
	Duration  int32     `json:"duration"`
	Budget    int32     `json:"budget"`
	StartDate time.Time `json:"start_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
