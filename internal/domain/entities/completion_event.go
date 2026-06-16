package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// CompletionProvider identifies which habit provider generated this completion.
type CompletionProvider string

const (
	ProviderLeetCode CompletionProvider = "leetcode"
	ProviderDuolingo CompletionProvider = "duolingo"
	ProviderCustom   CompletionProvider = "custom"
)

var validProviders = map[string]CompletionProvider{
	"leetcode": ProviderLeetCode,
	"duolingo": ProviderDuolingo,
	"custom":   ProviderCustom,
}

// CompletionEvent records a single completion of a HabitTask by a user.
// Raw provider payloads are stored in Metadata as JSONB (per D-01, D-02).
type CompletionEvent struct {
	ID          string
	HabitTaskID string
	UserID      string
	Provider    CompletionProvider
	OccurredAt  time.Time
	Metadata    datatypes.JSON
}

// NewCompletionEvent creates a validated CompletionEvent. provider must be one of
// "leetcode", "duolingo", or "custom"; otherwise an error is returned.
func NewCompletionEvent(habitTaskID, userID, provider string, occurredAt time.Time, metadata datatypes.JSON) (*CompletionEvent, error) {
	p, ok := validProviders[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	return &CompletionEvent{
		ID:          uuid.New().String(),
		HabitTaskID: habitTaskID,
		UserID:      userID,
		Provider:    p,
		OccurredAt:  occurredAt,
		Metadata:    metadata,
	}, nil
}
