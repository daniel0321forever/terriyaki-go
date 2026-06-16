package dto

import "time"

// HabitTaskDTO is the output DTO for a HabitTask entity.
type HabitTaskDTO struct {
	ID           string      `json:"id"`
	TaskType     string      `json:"taskType"`
	UserID       string      `json:"userID"`
	GrindID      string      `json:"grindID"`
	Date         time.Time   `json:"date"`
	FinishedTime *time.Time  `json:"finishedTime,omitempty"`
	Completed    bool        `json:"completed"`
	Metadata     interface{} `json:"metadata,omitempty"`
}
