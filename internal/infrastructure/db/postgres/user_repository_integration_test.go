//go:build integration
// +build integration

package postgres

import (
	"errors"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

func TestGormUserRepository_CreateAndFindByID(t *testing.T) {
	resetRepoTables(t)

	repo := NewGormUserRepository(Db)
	user, err := entities.NewUser("alice", "alice@example.com", "hashed-pass", "")
	if err != nil {
		t.Fatalf("failed to create user entity: %v", err)
	}

	err = repo.Create(user)
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	foundByID, err := repo.FindById(user.ID)
	if err != nil {
		t.Fatalf("find by id failed: %v", err)
	}
	if foundByID.Email != "alice@example.com" {
		t.Fatalf("expected normalized email alice@example.com, got %s", foundByID.Email)
	}

	foundByEmail, err := repo.FindByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("find by email failed: %v", err)
	}
	if foundByEmail.ID != user.ID {
		t.Fatalf("expected same user id, got %s and %s", foundByEmail.ID, user.ID)
	}
}

func TestGormUserRepository_FindByID_NotFound(t *testing.T) {
	resetRepoTables(t)

	repo := NewGormUserRepository(Db)
	_, err := repo.FindById("missing-id")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}
