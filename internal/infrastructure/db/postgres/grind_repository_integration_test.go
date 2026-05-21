//go:build integration
// +build integration

package postgres

import (
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

func TestGormGrindRepository_FindLatestByUserID(t *testing.T) {
	resetRepoTables(t)

	userRepo := NewGormUserRepository(Db)
	grindRepo := NewGormGrindRepository(Db)
	participationRepo := NewGormParticipationRepository(Db)

	user, err := entities.NewUser("bob", "bob@example.com", "hashed-pass", "")
	if err != nil {
		t.Fatalf("failed to create user entity: %v", err)
	}

	if err = userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	olderGrind, err := entities.NewGrind(7, 10, time.Now().UTC().AddDate(0, 0, -7))
	if err != nil {
		t.Fatalf("failed to create older grind entity: %v", err)
	}

	olderGrind.CreatedAt = time.Now().UTC().Add(-2 * time.Hour)
	olderGrind.UpdatedAt = olderGrind.CreatedAt

	if err = grindRepo.Create(olderGrind); err != nil {
		t.Fatalf("failed to persist older grind: %v", err)
	}

	newerGrind, err := entities.NewGrind(14, 20, time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create newer grind entity: %v", err)
	}

	newerGrind.CreatedAt = time.Now().UTC().Add(-1 * time.Hour)
	newerGrind.UpdatedAt = newerGrind.CreatedAt

	if err = grindRepo.Create(newerGrind); err != nil {
		t.Fatalf("failed to persist newer grind: %v", err)
	}

	olderParticipation, _ := entities.NewParticipation(user.ID, olderGrind.ID)
	if err = participationRepo.Create(olderParticipation); err != nil {
		t.Fatalf("failed to create older participation: %v", err)
	}

	newerParticipation, _ := entities.NewParticipation(user.ID, newerGrind.ID)
	if err = participationRepo.Create(newerParticipation); err != nil {
		t.Fatalf("failed to create newer participation: %v", err)
	}

	latest, err := grindRepo.FindLatestByUserID(user.ID)
	if err != nil {
		t.Fatalf("find latest grind failed: %v", err)
	}

	if latest.ID != newerGrind.ID {
		t.Fatalf("expected latest grind id %s, got %s", newerGrind.ID, latest.ID)
	}
}
