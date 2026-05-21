package mappers

import (
	"encoding/json"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildInterviewSessionDTO constructs InterviewSessionDTO from InterviewSession-related entity
func BuildInterviewSessionDTO(session *entities.InterviewSession) *dto.InterviewSessionDTO {
	var conversationHistory interface{}
	if session.ConversationHistory != nil {
		// datatypes.JSON is a type alias for []byte, so we can unmarshal it directly
		if err := json.Unmarshal(session.ConversationHistory, &conversationHistory); err != nil {
			// If unmarshalling fails, leave conversationHistory as nil
			conversationHistory = nil
		}
	}

	return &dto.InterviewSessionDTO{
		ID:                  session.ID,
		UserID:              session.UserID,
		TaskID:              session.TaskID,
		Status:              session.Status,
		ConversationHistory: conversationHistory,
		StartedAt:           session.StartedAt,
		EndedAt:             session.EndedAt,
	}
}
