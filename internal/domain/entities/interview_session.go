package entities

import (
    "errors"
    "time"
    // "github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type InterviewSession struct {
    gorm.Model
    ID                 string    `json:"id" gorm:"primaryKey"`
    UserID             string    `json:"user_id" gorm:"not null"`
    TaskID             string    `json:"task_id" gorm:"not null"`
    Status             string    `json:"status" gorm:"not null"` // "active", "completed", "paused"
    ConversationHistory datatypes.JSON `json:"conversation_history" gorm:"type:jsonb"`
    StartedAt          time.Time `json:"started_at" gorm:"not null"`
    EndedAt            *time.Time `json:"ended_at" gorm:""`
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
        ID:                 uuid.New().String(),
        UserID:             userID,
        TaskID:             taskID,
        Status:             "active",
        ConversationHistory: nil,
        StartedAt:          time.Now().UTC(),
        EndedAt:            nil,
    }, nil
}
