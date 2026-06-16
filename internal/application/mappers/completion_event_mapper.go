package mappers

import (
	"encoding/json"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildCompletionEventDTO constructs a CompletionEventDTO from a CompletionEvent entity.
func BuildCompletionEventDTO(event *entities.CompletionEvent) *dto.CompletionEventDTO {
	var metadata interface{}
	if event.Metadata != nil {
		if err := json.Unmarshal(event.Metadata, &metadata); err != nil {
			metadata = nil
		}
	}

	return &dto.CompletionEventDTO{
		ID:          event.ID,
		HabitTaskID: event.HabitTaskID,
		UserID:      event.UserID,
		Provider:    string(event.Provider),
		OccurredAt:  event.OccurredAt,
		Metadata:    metadata,
	}
}
