package dto

import "time"

// LeetCodeIngestPayload is the request body for leetcode ingestion.
type LeetCodeIngestPayload struct {
	GrindID           string    `json:"grindID"`
	ProblemTitle      string    `json:"problemTitle"`
	ProblemURL        string    `json:"problemURL"`
	ProblemDifficulty string    `json:"problemDifficulty"`
	ProblemTopicTags  []string  `json:"problemTopicTags"`
	Code              string    `json:"code"`
	CodeLanguage      string    `json:"codeLanguage"`
	OccurredAt        time.Time `json:"occurredAt"`
}

// DuolingoIngestPayload is the request body for duolingo ingestion.
type DuolingoIngestPayload struct {
	GrindID          string    `json:"grindID"`
	StreakCount      int       `json:"streakCount"`
	LessonsCompleted int       `json:"lessonsCompleted"`
	XPEarned         int       `json:"xpEarned"`
	OccurredAt       time.Time `json:"occurredAt"`
}

// CompletionEventDTO is the response DTO for a CompletionEvent entity.
type CompletionEventDTO struct {
	ID          string      `json:"id"`
	HabitTaskID string      `json:"habitTaskID"`
	UserID      string      `json:"userID"`
	Provider    string      `json:"provider"`
	OccurredAt  time.Time   `json:"occurredAt"`
	Metadata    interface{} `json:"metadata,omitempty"`
}
