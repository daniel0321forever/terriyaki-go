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
	ID                  string         `json:"id" gorm:"primaryKey"`
	UserID              string         `json:"user_id" gorm:"not null"`
	TaskID              string         `json:"task_id" gorm:"not null"`
	Status              string         `json:"status" gorm:"not null"` // "active", "completed", "paused"
	ConversationHistory datatypes.JSON `json:"conversation_history" gorm:"type:jsonb"`
	StartedAt           time.Time      `json:"started_at" gorm:"not null"`
	EndedAt             *time.Time     `json:"ended_at" gorm:""`

	// Interview recording + sharing (MVP starts with AudioRecordingURL)
	AudioRecordingURL      string     `json:"audio_recording_url" gorm:""`                   // URL/path to stored original audio
	IsShared               bool       `json:"is_shared" gorm:"default:false"`                // Whether shared to a grind
	SharedGrindID          string     `json:"shared_grind_id" gorm:""`                       // Grind it's shared to (if shared)
	ShareMode              string     `json:"share_mode" gorm:"default:'private'"`           // "private" | "public" | "anonymous"
	AnonymizedAudioURL     string     `json:"anonymized_audio_url" gorm:""`                  // URL/path to voice-converted audio
	VoiceConversionApplied bool       `json:"voice_conversion_applied" gorm:"default:false"` // Whether voice was converted
	AnonymizedVoiceID      string     `json:"anonymized_voice_id" gorm:""`                   // Voice ID used for conversion
	SharedAt               *time.Time `json:"shared_at" gorm:""`                             // When it was shared
}

func CreateInterviewSession(userID, taskID string) (*InterviewSession, error) {
	session := InterviewSession{
		ID:        uuid.New().String(),
		UserID:    userID,
		TaskID:    taskID,
		Status:    "active",
		StartedAt: time.Now(),
		ShareMode: "private",
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
