package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// IngestionProvider defines the interface for habit-source-specific payload parsing.
type IngestionProvider interface {
	ParsePayload(raw map[string]interface{}) (*ingestResult, error)
	ProviderName() string
}

// ingestResult holds the parsed output from an IngestionProvider before entity creation.
type ingestResult struct {
	Provider   entities.CompletionProvider
	OccurredAt time.Time
	Metadata   datatypes.JSON
}

// LeetCodeProvider parses LeetCode Chrome-extension payloads.
type LeetCodeProvider struct{}

func (p *LeetCodeProvider) ProviderName() string { return "leetcode" }

func (p *LeetCodeProvider) ParsePayload(raw map[string]interface{}) (*ingestResult, error) {
	metaBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal leetcode payload: %w", err)
	}

	occurredAt := time.Now().UTC()
	if v, ok := raw["occurredAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			occurredAt = t
		}
	}

	return &ingestResult{
		Provider:   entities.ProviderLeetCode,
		OccurredAt: occurredAt,
		Metadata:   datatypes.JSON(metaBytes),
	}, nil
}

// DuolingoProvider parses Duolingo webhook payloads.
type DuolingoProvider struct{}

func (p *DuolingoProvider) ProviderName() string { return "duolingo" }

func (p *DuolingoProvider) ParsePayload(raw map[string]interface{}) (*ingestResult, error) {
	metaBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal duolingo payload: %w", err)
	}

	occurredAt := time.Now().UTC()
	if v, ok := raw["occurredAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			occurredAt = t
		}
	}

	return &ingestResult{
		Provider:   entities.ProviderDuolingo,
		OccurredAt: occurredAt,
		Metadata:   datatypes.JSON(metaBytes),
	}, nil
}

// IngestService orchestrates ingestion of completion events from external providers.
type IngestService struct {
	habitTaskRepo       repositories.HabitTaskRepository
	completionEventRepo repositories.CompletionEventRepository
	providers           map[string]IngestionProvider
}

// NewIngestService constructs an IngestService with LeetCode and Duolingo providers registered.
func NewIngestService(
	habitTaskRepo repositories.HabitTaskRepository,
	completionEventRepo repositories.CompletionEventRepository,
) *IngestService {
	return &IngestService{
		habitTaskRepo:       habitTaskRepo,
		completionEventRepo: completionEventRepo,
		providers: map[string]IngestionProvider{
			"leetcode": &LeetCodeProvider{},
			"duolingo": &DuolingoProvider{},
		},
	}
}

// Ingest validates the provider, finds today's habit task, parses the payload, and
// persists a CompletionEvent. Returns ErrHabitTaskNotFound when no task exists today.
func (s *IngestService) Ingest(providerName, userID, grindID string, rawPayload map[string]interface{}) (*entities.CompletionEvent, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	todayTask, err := s.habitTaskRepo.FindTodayTask(userID, grindID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, config.ErrHabitTaskNotFound
		}
		return nil, fmt.Errorf("failed to find today's habit task: %w", err)
	}

	result, err := provider.ParsePayload(rawPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payload: %w", err)
	}

	event, err := entities.NewCompletionEvent(
		todayTask.ID,
		userID,
		string(result.Provider),
		result.OccurredAt,
		result.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build completion event: %w", err)
	}

	if err := s.completionEventRepo.Create(event); err != nil {
		return nil, fmt.Errorf("failed to persist completion event: %w", err)
	}

	return event, nil
}
