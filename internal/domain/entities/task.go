package entities

import (
	"time"

	// "github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// TODO: this is not general enough and only designed for leetcode
type Task struct {
	ID                 string
	TaskType           string
	UserID             string
	GrindID            string
	Date               time.Time
	FinishedTime       time.Time
	Completed          bool
	ProblemTitle       *string
	ProblemDescription *string
	ProblemURL         *string
	ProblemDifficulty  *string
	ProblemTopicTags   datatypes.JSON

	Code         string
	CodeLanguage string
}

/** Constructor in factory pattern
 */
func NewTask(userID string, grindID string, date time.Time) (*Task, error) {
	return &Task{
		ID:        uuid.New().String(),
		TaskType:  "leetcode",
		UserID:    userID,
		GrindID:   grindID,
		Date:      date,
		Completed: false,
	}, nil
}

// Logic: Check if a problem is assigned
func (t *Task) HasProblemAssigned() bool {
	return t.ProblemTitle != nil
}
