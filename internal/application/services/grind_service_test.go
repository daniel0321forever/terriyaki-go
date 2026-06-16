package services

import (
	"errors"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
)

// NOTE: Tests for transactional write methods (CreateGroupGrind, DeleteGrind,
// AddParticipation write paths, AcceptInvitation) are covered by integration tests in
// grind_repository_integration_test.go which use a real Postgres testcontainer.
// Unit tests here only cover guard/validation logic that executes BEFORE the transaction begins.

func TestGrindServiceGetOngoingGrindByUserID_EndedGrind(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindLatestByUserID", "u1").Return(&entities.Grind{
		ID:        "g1",
		Duration:  5,
		StartDate: time.Now().UTC().AddDate(0, 0, -10),
	}, nil)

	svc := NewGrindService(nil, grindRepo, userRepo, habitTaskRepo, partRepo, msgRepo)

	_, err := svc.GetOngoingGrindByUserID(dto.GetOngoingGrindDTO{UserID: "u1"})
	if !errors.Is(err, config.ErrNoOngoingGrind) {
		t.Fatalf("expected ErrNoOngoingGrind, got %v", err)
	}
}

func TestGrindServiceGetOngoingGrindByUserID_GrindNotFound(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindLatestByUserID", "u1").Return(nil, errors.New("missing"))

	svc := NewGrindService(nil, grindRepo, userRepo, habitTaskRepo, partRepo, msgRepo)

	_, err := svc.GetOngoingGrindByUserID(dto.GetOngoingGrindDTO{UserID: "u1"})
	if !errors.Is(err, config.ErrGrindNotFound) {
		t.Fatalf("expected ErrGrindNotFound, got %v", err)
	}
}

func TestGrindServiceGetOngoingGrindByUserID_UserQuitted(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindLatestByUserID", "u1").Return(&entities.Grind{
		ID:        "g1",
		Duration:  5,
		StartDate: time.Now().UTC().AddDate(0, 0, -1),
	}, nil)
	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(&entities.Participation{UserID: "u1", GrindID: "g1", Quitted: true}, nil)

	svc := NewGrindService(nil, grindRepo, userRepo, habitTaskRepo, partRepo, msgRepo)

	_, err := svc.GetOngoingGrindByUserID(dto.GetOngoingGrindDTO{UserID: "u1"})
	if !errors.Is(err, config.ErrUserNotParticipatingOrQuit) {
		t.Fatalf("expected ErrUserNotParticipatingOrQuit, got %v", err)
	}
}

// AddParticipation guard tests — these check validation BEFORE the transaction opens.

func TestGrindServiceAddParticipation_AlreadyExists(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(&entities.Participation{ID: "p1", UserID: "u1", GrindID: "g1"}, nil)

	svc := NewGrindService(nil, grindRepo, userRepo, habitTaskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if err == nil {
		t.Fatalf("expected already exists error, got nil")
	}
}

func TestGrindServiceAddParticipation_UserNotFound(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(nil, nil)
	userRepo.On("FindById", "u1").Return(nil, errors.New("missing"))

	svc := NewGrindService(nil, grindRepo, userRepo, habitTaskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if !errors.Is(err, config.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGrindServiceAddParticipation_GrindNotFound(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	habitTaskRepo := new(mocks.MockHabitTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(nil, nil)
	userRepo.On("FindById", "u1").Return(&entities.User{ID: "u1"}, nil)
	grindRepo.On("FindById", "g1").Return(nil, errors.New("missing"))

	svc := NewGrindService(nil, grindRepo, userRepo, habitTaskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if !errors.Is(err, config.ErrGrindNotFound) {
		t.Fatalf("expected ErrGrindNotFound, got %v", err)
	}
}
