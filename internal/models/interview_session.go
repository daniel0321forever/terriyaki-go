package models

import (
    "time"
    "github.com/daniel0321forever/terriyaki-go/internal/database"
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

func CreateInterviewSession(userID, taskID string) (*InterviewSession, error) {
    session := InterviewSession{
        ID:        uuid.New().String(),
        UserID:    userID,
        TaskID:    taskID,
        Status:    "active",
        StartedAt: time.Now(),
    }

    result := database.Db.Create(&session)
    if result.Error != nil {
        return nil, result.Error
    }
    return &session, nil
}

func GetInterviewSession(sessionID string) (*InterviewSession, error) {
    var session InterviewSession
    result := database.Db.Where("id = ?", sessionID).First(&session)
    if result.Error != nil {
        return nil, result.Error
    }
    return &session, nil
}