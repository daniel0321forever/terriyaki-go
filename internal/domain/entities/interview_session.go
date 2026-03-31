package entities

import (
	"errors"
	"time"

	// "github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// specific to code interview for now
type InterviewSession struct {
	ID                  string
	UserID              string
	TaskID              string
	Status              string // "active", "completed", "paused"
	ConversationHistory datatypes.JSON
	StartedAt           time.Time
	EndedAt             *time.Time
}

/** Constructor in factory pattern
 * @param userID - the ID of the user starting the interview
 * @param taskID - the ID of the task being interviewed on
 * @return the created interview session
 */
func NewInterviewSession(userID, taskID string) (*InterviewSession, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}
	if taskID == "" {
		return nil, errors.New("taskID cannot be empty")
	}

	return &InterviewSession{
		ID:                  uuid.New().String(),
		UserID:              userID,
		TaskID:              taskID,
		Status:              "active",
		ConversationHistory: nil,
		StartedAt:           time.Now().UTC(),
		EndedAt:             nil,
	}, nil
}
