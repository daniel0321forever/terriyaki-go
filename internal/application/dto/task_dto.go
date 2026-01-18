package dto

import "time"

// Input DTOs
type GetTaskDTO struct {
	TaskID             string
	SetProblemIfNeeded bool
}

type GetTodayTaskDTO struct {
	UserID  string
	GrindID string
}

type FinishTaskDTO struct {
	TaskID       string
	Code         string
	CodeLanguage string
}

type GetCompletionStatsDTO struct {
	GrindID string
}

// Output DTOs
type TaskDTO struct {
	ID                 string      `json:"id"`
	TaskType           string      `json:"type"`
	UserID             string      `json:"userID"`
	GrindID            string      `json:"grindID"`
	Date               time.Time   `json:"date"`
	FinishedTime       time.Time   `json:"finishedTime"`
	Completed          bool        `json:"completed"`
	ProblemTitle       *string     `json:"title,omitempty"`
	ProblemDescription *string     `json:"description,omitempty"`
	ProblemURL         *string     `json:"url,omitempty"`
	ProblemDifficulty  *string     `json:"difficulty,omitempty"`
	ProblemTopicTags   interface{} `json:"topicTags,omitempty"`
	Code               string      `json:"code"`
	CodeLanguage       string      `json:"language"`
}
