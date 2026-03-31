package entities

import (
	"strings"
	"testing"
	"time"
)

func TestNewGrind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		duration    int
		budget      int
		startDate   time.Time
		wantErr     bool
		errContains string
	}{
		{
			name:      "creates grind with valid inputs",
			duration:  30,
			budget:    500,
			startDate: time.Date(2026, 3, 29, 15, 4, 5, 0, time.FixedZone("UTC+8", 8*60*60)),
		},
		{
			name:        "rejects duration less than one day",
			duration:    0,
			budget:      100,
			startDate:   time.Now(),
			wantErr:     true,
			errContains: "duration must be at least 1 day",
		},
		{
			name:        "rejects negative budget",
			duration:    10,
			budget:      -1,
			startDate:   time.Now(),
			wantErr:     true,
			errContains: "budget cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grind, err := NewGrind(tt.duration, tt.budget, tt.startDate)

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

			if grind == nil {
				t.Fatalf("expected grind, got nil")
			}

			if grind.ID == "" {
				t.Fatalf("expected non-empty ID")
			}

			if grind.Budget != int32(tt.budget) {
				t.Fatalf("expected budget %d, got %d", tt.budget, grind.Budget)
			}

			if grind.Duration != int32(tt.duration) {
				t.Fatalf("expected duration %d, got %d", tt.duration, grind.Duration)
			}

			if len(grind.Participants) != 0 {
				t.Fatalf("expected empty participants, got %d", len(grind.Participants))
			}

			if len(grind.Tasks) != 0 {
				t.Fatalf("expected empty tasks, got %d", len(grind.Tasks))
			}

			if grind.StartDate.Location() != time.UTC {
				t.Fatalf("expected start date in UTC, got %v", grind.StartDate.Location())
			}

			if grind.CreatedAt.IsZero() || grind.UpdatedAt.IsZero() {
				t.Fatalf("expected created/updated timestamps to be set")
			}

			if grind.CreatedAt.Location() != time.UTC || grind.UpdatedAt.Location() != time.UTC {
				t.Fatalf("expected created/updated timestamps to be UTC")
			}
		})
	}
}
