package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewCompletionEvent_ValidLeetcode(t *testing.T) {
	now := time.Now()
	event, err := NewCompletionEvent("habit-1", "user-1", "leetcode", now, nil)
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, "habit-1", event.HabitTaskID)
	assert.Equal(t, "user-1", event.UserID)
	assert.Equal(t, ProviderLeetCode, event.Provider)
	assert.Nil(t, event.Metadata)
}

func Test_NewCompletionEvent_UnknownProvider(t *testing.T) {
	now := time.Now()
	event, err := NewCompletionEvent("habit-1", "user-1", "unknown", now, nil)
	require.Error(t, err)
	assert.Nil(t, event)
	assert.Equal(t, "unsupported provider: unknown", err.Error())
}
