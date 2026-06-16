package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	task, err := NewTask("user-1", "grind-1", date)
	require.NoError(t, err)
	require.NotNil(t, task)

	if task.ID == "" {
		t.Fatalf("expected non-empty task ID")
	}
	if task.TaskType != "leetcode" {
		t.Fatalf("expected default task type leetcode, got %q", task.TaskType)
	}
	if task.UserID != "user-1" {
		t.Fatalf("expected userID user-1, got %q", task.UserID)
	}
	if task.GrindID != "grind-1" {
		t.Fatalf("expected grindID grind-1, got %q", task.GrindID)
	}
	if !task.Date.Equal(date) {
		t.Fatalf("expected date %v, got %v", date, task.Date)
	}
	if task.Completed {
		t.Fatalf("expected new task to be incomplete")
	}
}

// TODO: // LeetCode-specific fields
func TestTaskHasProblemAssigned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		problemTitle  *string
		expectedValue bool
	}{
		{name: "no problem title", problemTitle: nil, expectedValue: false},
		{name: "problem title assigned", problemTitle: strPtr("Two Sum"), expectedValue: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{ProblemTitle: tt.problemTitle}
			if got := task.HasProblemAssigned(); got != tt.expectedValue {
				t.Fatalf("expected %v, got %v", tt.expectedValue, got)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
