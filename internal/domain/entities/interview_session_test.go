package entities

import (
	"strings"
	"testing"
	"time"
)

func TestNewInterviewSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userID      string
		taskID      string
		wantErr     bool
		errContains string
	}{
		{name: "creates active session", userID: "user-1", taskID: "task-1"},
		{name: "rejects empty userID", userID: "", taskID: "task-1", wantErr: true, errContains: "userID cannot be empty"},
		{name: "rejects empty taskID", userID: "user-1", taskID: "", wantErr: true, errContains: "taskID cannot be empty"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewInterviewSession(tt.userID, tt.taskID)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if s == nil {
				t.Fatalf("expected session, got nil")
			}
			if s.ID == "" {
				t.Fatalf("expected non-empty session ID")
			}
			if s.UserID != tt.userID {
				t.Fatalf("expected userID %q, got %q", tt.userID, s.UserID)
			}
			if s.TaskID != tt.taskID {
				t.Fatalf("expected taskID %q, got %q", tt.taskID, s.TaskID)
			}
			if s.Status != "active" {
				t.Fatalf("expected status active, got %q", s.Status)
			}
			if s.EndedAt != nil {
				t.Fatalf("expected EndedAt nil for new session")
			}
			if s.StartedAt.IsZero() {
				t.Fatalf("expected StartedAt to be set")
			}
			if s.StartedAt.Location() != time.UTC {
				t.Fatalf("expected StartedAt to be UTC")
			}
		})
	}
}
