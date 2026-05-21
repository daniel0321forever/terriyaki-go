package dto

import "time"

// Input DTOs
type CreateInterviewSessionDTO struct {
	UserID string
	TaskID string
}

type GetInterviewSessionDTO struct {
	SessionID string
}

type UpdateInterviewSessionDTO struct {
	SessionID           string
	Status              *string     `json:"status,omitempty"` // "active", "completed", "paused"
	ConversationHistory interface{} `json:"conversationHistory,omitempty"`
	EndedAt             *time.Time  `json:"endedAt,omitempty"`
}

// Output DTOs
type InterviewSessionDTO struct {
	ID                  string      `json:"id"`
	UserID              string      `json:"userID"`
	TaskID              string      `json:"taskID"`
	Status              string      `json:"status"` // "active", "completed", "paused"
	ConversationHistory interface{} `json:"conversationHistory,omitempty"`
	StartedAt           time.Time   `json:"startedAt"`
	EndedAt             *time.Time  `json:"endedAt,omitempty"`
}
