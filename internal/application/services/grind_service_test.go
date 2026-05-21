package services

import (
	"errors"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
	"github.com/stretchr/testify/mock"
)

func TestGrindServiceGetOngoingGrindByUserID_EndedGrind(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindLatestByUserID", "u1").Return(&entities.Grind{
		ID:        "g1",
		Duration:  5,
		StartDate: time.Now().UTC().AddDate(0, 0, -10),
	}, nil)

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)

	_, err := svc.GetOngoingGrindByUserID(dto.GetOngoingGrindDTO{UserID: "u1"})
	if !errors.Is(err, config.ErrNoOngoingGrind) {
		t.Fatalf("expected ErrNoOngoingGrind, got %v", err)
	}
}

func TestGrindServiceGetOngoingGrindByUserID_GrindNotFound(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindLatestByUserID", "u1").Return(nil, errors.New("missing"))

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)

	_, err := svc.GetOngoingGrindByUserID(dto.GetOngoingGrindDTO{UserID: "u1"})
	if !errors.Is(err, config.ErrGrindNotFound) {
		t.Fatalf("expected ErrGrindNotFound, got %v", err)
	}
}

func TestGrindServiceGetOngoingGrindByUserID_UserQuitted(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindLatestByUserID", "u1").Return(&entities.Grind{
		ID:        "g1",
		Duration:  5,
		StartDate: time.Now().UTC().AddDate(0, 0, -1),
	}, nil)
	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(&entities.Participation{UserID: "u1", GrindID: "g1", Quitted: true}, nil)

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)

	_, err := svc.GetOngoingGrindByUserID(dto.GetOngoingGrindDTO{UserID: "u1"})
	if !errors.Is(err, config.ErrUserNotParticipatingOrQuit) {
		t.Fatalf("expected ErrUserNotParticipatingOrQuit, got %v", err)
	}
}

func TestGrindServiceAddParticipation_AlreadyExists(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(&entities.Participation{ID: "p1", UserID: "u1", GrindID: "g1"}, nil)

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if err == nil {
		t.Fatalf("expected already exists error, got nil")
	}
}

func TestGrindServiceAddParticipation_UserNotFound(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(nil, nil)
	userRepo.On("FindById", "u1").Return(nil, errors.New("missing"))

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if !errors.Is(err, config.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGrindServiceAddParticipation_GrindNotFound(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(nil, nil)
	userRepo.On("FindById", "u1").Return(&entities.User{ID: "u1"}, nil)
	grindRepo.On("FindById", "g1").Return(nil, errors.New("missing"))

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if !errors.Is(err, config.ErrGrindNotFound) {
		t.Fatalf("expected ErrGrindNotFound, got %v", err)
	}
}

func TestGrindServiceAddParticipation_CreateParticipationFailure(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindById", "g1").Return(&entities.Grind{ID: "g1", Duration: 3, StartDate: time.Now().UTC()}, nil)
	userRepo.On("FindById", "u1").Return(&entities.User{ID: "u1"}, nil)
	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(nil, nil)
	partRepo.On("Create", mock.MatchedBy(func(p *entities.Participation) bool {
		return p.UserID == "u1" && p.GrindID == "g1"
	})).Return(errors.New("create failed"))

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if err == nil || err.Error() != "create failed" {
		t.Fatalf("expected create failed, got %v", err)
	}
	taskRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestGrindServiceAddParticipation_TaskCreateFailure(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindById", "g1").Return(&entities.Grind{ID: "g1", Duration: 3, StartDate: time.Now().UTC()}, nil)
	userRepo.On("FindById", "u1").Return(&entities.User{ID: "u1"}, nil)
	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(nil, nil)
	partRepo.On("Create", mock.AnythingOfType("*entities.Participation")).Return(nil)
	taskRepo.On("Create", mock.AnythingOfType("*entities.Task")).Return(errors.New("task create failed")).Once()

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)
	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if err == nil || err.Error() != "task create failed" {
		t.Fatalf("expected task create failed, got %v", err)
	}
	taskRepo.AssertNumberOfCalls(t, "Create", 1)
}

func TestGrindServiceAddParticipation_SuccessCreatesTasks(t *testing.T) {
	t.Parallel()

	grind := &entities.Grind{ID: "g1", Duration: 3, StartDate: time.Now().UTC()}

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	grindRepo.On("FindById", "g1").Return(grind, nil)
	userRepo.On("FindById", "u1").Return(&entities.User{ID: "u1"}, nil)
	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(nil, nil)
	partRepo.On("Create", mock.MatchedBy(func(p *entities.Participation) bool {
		return p.UserID == "u1" && p.GrindID == "g1"
	})).Return(nil)
	taskRepo.On("Create", mock.MatchedBy(func(t *entities.Task) bool {
		return t.UserID == "u1" && t.GrindID == "g1"
	})).Return(nil)

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)

	err := svc.AddParticipation(dto.AddParticipationDTO{UserID: "u1", GrindID: "g1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	taskRepo.AssertNumberOfCalls(t, "Create", 3)
}

func TestGrindServiceDeleteGrind_TaskDeletionFailure(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	userRepo := new(mocks.MockUserRepository)
	taskRepo := new(mocks.MockTaskRepository)
	partRepo := new(mocks.MockParticipationRepository)
	msgRepo := new(mocks.MockMessageRepository)

	taskRepo.On("DeleteByGrindID", "g1").Return(errors.New("task delete fail"))

	svc := NewGrindService(grindRepo, userRepo, taskRepo, partRepo, msgRepo)

	err := svc.DeleteGrind(dto.DeleteGrindDTO{GrindID: "g1"})
	if err == nil || err.Error() != "task delete fail" {
		t.Fatalf("expected task delete fail, got %v", err)
	}
}
