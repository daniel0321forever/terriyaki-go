package entities

import "testing"

func TestNewParticipation(t *testing.T) {
	t.Parallel()

	p, err := NewParticipation("user-1", "grind-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatalf("expected participation, got nil")
	}
	if p.ID == "" {
		t.Fatalf("expected non-empty participation ID")
	}
	if p.UserID != "user-1" {
		t.Fatalf("expected userID user-1, got %q", p.UserID)
	}
	if p.GrindID != "grind-1" {
		t.Fatalf("expected grindID grind-1, got %q", p.GrindID)
	}
	if p.MissedDays != 0 {
		t.Fatalf("expected MissedDays = 0, got %d", p.MissedDays)
	}
	if p.TotalPenalty != 0 {
		t.Fatalf("expected TotalPenalty = 0, got %d", p.TotalPenalty)
	}
	if p.Quitted {
		t.Fatalf("expected Quitted = false")
	}
	if !p.QuittedAt.IsZero() {
		t.Fatalf("expected zero QuittedAt for new participation")
	}
}
