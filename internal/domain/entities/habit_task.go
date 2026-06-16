package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// HabitTask is the generic habit activity entity.
// Provider-specific fields (e.g. LeetCode problem title, Duolingo lesson) are
// stored in Metadata as JSONB rather than as typed struct fields (per D-01).
type HabitTask struct {
	ID           string
	TaskType     string
	UserID       string
	GrindID      string
	Date         time.Time
	FinishedTime *time.Time
	Completed    bool
	Metadata     datatypes.JSON
}

// NewHabitTask creates a new HabitTask with a generated UUID, TaskType="generic",
// and Completed=false. userID and grindID must be non-empty.
func NewHabitTask(userID, grindID string, date time.Time) (*HabitTask, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}
	if grindID == "" {
		return nil, errors.New("grindID cannot be empty")
	}
	return &HabitTask{
		ID:        uuid.New().String(),
		TaskType:  "generic",
		UserID:    userID,
		GrindID:   grindID,
		Date:      date,
		Completed: false,
		Metadata:  nil,
	}, nil
}
