//go:build integration
// +build integration

package postgres_test

import (
	"errors"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"gorm.io/gorm"
)

func TestGormGrindRepository_FindLatestByUserID(t *testing.T) {
	resetRepoTables(t)

	userRepo := postgres.NewGormUserRepository(postgres.Db)
	grindRepo := postgres.NewGormGrindRepository(postgres.Db)
	participationRepo := postgres.NewGormParticipationRepository(postgres.Db)

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

// TestCreateGroupGrindRollback verifies that if habitTaskRepo.Create fails mid-loop,
// the grind and participation rows written before the failure are rolled back.
func TestCreateGroupGrindRollback(t *testing.T) {
	resetRepoTables(t)

	userRepo := postgres.NewGormUserRepository(postgres.Db)

	user, err := entities.NewUser("grindrollback", "grindrollback@example.com", "hashed-pass", "")
	if err != nil {
		t.Fatalf("failed to create user entity: %v", err)
	}
	if err = userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Use a failing habitTaskRepo that fails immediately on the first Create call
	failingHabitTaskRepo := &failAfterNHabitTaskRepo{
		inner:      postgres.NewGormHabitTaskRepository(postgres.Db),
		failAfterN: 0,
	}

	grindService := services.NewGrindService(
		postgres.Db,
		postgres.NewGormGrindRepository(postgres.Db),
		userRepo,
		failingHabitTaskRepo,
		postgres.NewGormParticipationRepository(postgres.Db),
		postgres.NewGormMessageRepository(postgres.Db),
	)

	startDate := time.Now().UTC()
	_, err = grindService.CreateGroupGrind(dto.CreateGrindDTO{
		Duration:  3,
		Budget:    100,
		StartDate: startDate,
		CreatorID: user.ID,
	})
	if err == nil {
		t.Fatal("expected error from CreateGroupGrind due to failing habitTaskRepo, got nil")
	}

	// Assert no grind rows remain
	var grindCount int64
	postgres.Db.Table("grinds").Count(&grindCount)
	if grindCount != 0 {
		t.Fatalf("expected 0 grind rows after rollback, got %d", grindCount)
	}

	// Assert no participation rows remain
	var participationCount int64
	postgres.Db.Table("participation").Count(&participationCount)
	if participationCount != 0 {
		t.Fatalf("expected 0 participation rows after rollback, got %d", participationCount)
	}

	// Assert no habit_tasks rows remain
	var habitTaskCount int64
	postgres.Db.Table("habit_tasks").Count(&habitTaskCount)
	if habitTaskCount != 0 {
		t.Fatalf("expected 0 habit_task rows after rollback, got %d", habitTaskCount)
	}
}

// TestDeleteGrindRollback verifies that if participationRepo.DeleteByGrindID fails,
// the habitTask deletions that were supposed to happen together are rolled back
// (habitTask rows still exist after the failed delete attempt).
func TestDeleteGrindRollback(t *testing.T) {
	resetRepoTables(t)

	userRepo := postgres.NewGormUserRepository(postgres.Db)
	grindRepo := postgres.NewGormGrindRepository(postgres.Db)
	participationRepo := postgres.NewGormParticipationRepository(postgres.Db)
	habitTaskRepo := postgres.NewGormHabitTaskRepository(postgres.Db)

	// Create a user and a grind with tasks
	user, _ := entities.NewUser("deleterollback", "deleterollback@example.com", "hashed-pass", "")
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	grind, _ := entities.NewGrind(2, 50, time.Now().UTC())
	if err := grindRepo.Create(grind); err != nil {
		t.Fatalf("failed to create grind: %v", err)
	}

	participation, _ := entities.NewParticipation(user.ID, grind.ID)
	if err := participationRepo.Create(participation); err != nil {
		t.Fatalf("failed to create participation: %v", err)
	}

	task1, _ := entities.NewHabitTask(user.ID, grind.ID, time.Now().UTC())
	task2, _ := entities.NewHabitTask(user.ID, grind.ID, time.Now().UTC().AddDate(0, 0, 1))
	if err := habitTaskRepo.Create(task1); err != nil {
		t.Fatalf("failed to create habit task 1: %v", err)
	}
	if err := habitTaskRepo.Create(task2); err != nil {
		t.Fatalf("failed to create habit task 2: %v", err)
	}

	// Count habit_tasks before delete attempt
	var beforeCount int64
	postgres.Db.Table("habit_tasks").Where("grind_id = ?", grind.ID).Count(&beforeCount)
	if beforeCount != 2 {
		t.Fatalf("expected 2 habit_task rows before delete, got %d", beforeCount)
	}

	// Use a failing participationRepo that always fails on DeleteByGrindID
	failingPartRepo := &failingParticipationRepo{inner: participationRepo}

	grindService := services.NewGrindService(
		postgres.Db,
		grindRepo,
		userRepo,
		habitTaskRepo,
		failingPartRepo,
		postgres.NewGormMessageRepository(postgres.Db),
	)

	err := grindService.DeleteGrind(dto.DeleteGrindDTO{GrindID: grind.ID})
	if err == nil {
		t.Fatal("expected error from DeleteGrind due to failing participationRepo, got nil")
	}

	// Assert habit_task rows still exist (rollback worked)
	var afterCount int64
	postgres.Db.Table("habit_tasks").Where("grind_id = ?", grind.ID).Count(&afterCount)
	if afterCount != 2 {
		t.Fatalf("expected 2 habit_task rows after rollback, got %d", afterCount)
	}
}

// TestAcceptInvitationRollback verifies that if the message Update fails inside
// AcceptInvitation, any participation rows created within the transaction are rolled back.
func TestAcceptInvitationRollback(t *testing.T) {
	resetRepoTables(t)

	userRepo := postgres.NewGormUserRepository(postgres.Db)
	grindRepo := postgres.NewGormGrindRepository(postgres.Db)
	participationRepo := postgres.NewGormParticipationRepository(postgres.Db)
	habitTaskRepo := postgres.NewGormHabitTaskRepository(postgres.Db)
	messageRepo := postgres.NewGormMessageRepository(postgres.Db)

	// Create invitor and accepter users
	invitor, _ := entities.NewUser("invitor", "invitor@example.com", "hashed-pass", "")
	if err := userRepo.Create(invitor); err != nil {
		t.Fatalf("failed to create invitor user: %v", err)
	}

	accepter, _ := entities.NewUser("accepter", "accepter@example.com", "hashed-pass", "")
	if err := userRepo.Create(accepter); err != nil {
		t.Fatalf("failed to create accepter user: %v", err)
	}

	// Create a grind (invitor is creator)
	grind, _ := entities.NewGrind(2, 50, time.Now().UTC())
	if err := grindRepo.Create(grind); err != nil {
		t.Fatalf("failed to create grind: %v", err)
	}

	// Create invitor participation
	invitorParticipation, _ := entities.NewParticipation(invitor.ID, grind.ID)
	if err := participationRepo.Create(invitorParticipation); err != nil {
		t.Fatalf("failed to create invitor participation: %v", err)
	}

	// Create invitation message
	inviteMsg, _ := entities.NewMessage(
		invitor.ID, accepter.ID,
		"join my grind", "invitation",
		grind.ID, false, false,
	)
	if err := messageRepo.Create(inviteMsg); err != nil {
		t.Fatalf("failed to create invitation message: %v", err)
	}

	// Count participation before the accept attempt
	var beforeCount int64
	postgres.Db.Table("participation").Count(&beforeCount)
	if beforeCount != 1 {
		t.Fatalf("expected 1 participation row before accept, got %d", beforeCount)
	}

	// Use a failing messageRepo that always fails on Update
	failingMsgRepo := &failingMessageRepo{inner: messageRepo}

	grindService := services.NewGrindService(
		postgres.Db,
		grindRepo,
		userRepo,
		habitTaskRepo,
		participationRepo,
		postgres.NewGormMessageRepository(postgres.Db),
	)

	err := grindService.AcceptInvitation(
		dto.AddParticipationDTO{GrindID: grind.ID, UserID: accepter.ID},
		dto.UpdateMessageInvitationAcceptedStatusDTO{MessageID: inviteMsg.ID, Accepted: true},
		dto.CreateInvitationAcceptedMessageDTO{AccepterID: accepter.ID, InvitorID: invitor.ID, GrindID: grind.ID},
		failingMsgRepo,
	)
	if err == nil {
		t.Fatal("expected error from AcceptInvitation due to failing messageRepo.Update, got nil")
	}

	// Assert participation count is still 1 (accepter's row was rolled back)
	var afterCount int64
	postgres.Db.Table("participation").Count(&afterCount)
	if afterCount != 1 {
		t.Fatalf("expected 1 participation row after rollback (only invitor), got %d", afterCount)
	}
}

// --- Test doubles ---

// failAfterNHabitTaskRepo wraps GormHabitTaskRepository and fails on Create after N successful calls.
type failAfterNHabitTaskRepo struct {
	inner      *postgres.GormHabitTaskRepository
	failAfterN int
	callCount  int
}

func (r *failAfterNHabitTaskRepo) WithTx(tx *gorm.DB) repositories.HabitTaskRepository {
	return &failAfterNHabitTaskRepo{
		inner:      r.inner.WithTx(tx).(*postgres.GormHabitTaskRepository),
		failAfterN: r.failAfterN,
		callCount:  r.callCount,
	}
}

func (r *failAfterNHabitTaskRepo) Create(task *entities.HabitTask) error {
	if r.callCount >= r.failAfterN {
		return errors.New("injected habitTaskRepo.Create failure")
	}
	r.callCount++
	return r.inner.Create(task)
}

func (r *failAfterNHabitTaskRepo) FindByID(id string) (*entities.HabitTask, error) {
	return r.inner.FindByID(id)
}

func (r *failAfterNHabitTaskRepo) FindByGrindIDAndUserID(grindID, userID string) ([]*entities.HabitTask, error) {
	return r.inner.FindByGrindIDAndUserID(grindID, userID)
}

func (r *failAfterNHabitTaskRepo) FindTodayTask(userID, grindID string) (*entities.HabitTask, error) {
	return r.inner.FindTodayTask(userID, grindID)
}

func (r *failAfterNHabitTaskRepo) Update(task *entities.HabitTask) error {
	return r.inner.Update(task)
}

func (r *failAfterNHabitTaskRepo) FindByGrindIDAndParticipantID(grindID, participantID string) ([]entities.HabitTask, error) {
	return r.inner.FindByGrindIDAndParticipantID(grindID, participantID)
}

func (r *failAfterNHabitTaskRepo) DeleteByGrindID(grindID string) error {
	return r.inner.DeleteByGrindID(grindID)
}

// failingParticipationRepo always fails on DeleteByGrindID.
type failingParticipationRepo struct {
	inner *postgres.GormParticipationRepository
}

func (r *failingParticipationRepo) WithTx(tx *gorm.DB) repositories.ParticipationRepository {
	return &failingParticipationRepo{
		inner: r.inner.WithTx(tx).(*postgres.GormParticipationRepository),
	}
}

func (r *failingParticipationRepo) Create(p *entities.Participation) error {
	return r.inner.Create(p)
}

func (r *failingParticipationRepo) FindByParticipationId(id string) (*entities.Participation, error) {
	return r.inner.FindByParticipationId(id)
}

func (r *failingParticipationRepo) FindByUserAndGrind(userID, grindID string) (*entities.Participation, error) {
	return r.inner.FindByUserAndGrind(userID, grindID)
}

func (r *failingParticipationRepo) Update(p *entities.Participation) error {
	return r.inner.Update(p)
}

func (r *failingParticipationRepo) DeleteByGrindID(grindID string) error {
	return errors.New("injected participationRepo.DeleteByGrindID failure")
}

// failingMessageRepo always fails on Update.
type failingMessageRepo struct {
	inner *postgres.GormMessageRepository
}

func (r *failingMessageRepo) WithTx(tx *gorm.DB) repositories.MessageRepository {
	return &failingMessageRepo{
		inner: r.inner.WithTx(tx).(*postgres.GormMessageRepository),
	}
}

func (r *failingMessageRepo) Create(m *entities.Message) error {
	return r.inner.Create(m)
}

func (r *failingMessageRepo) FindByID(id string) (*entities.Message, error) {
	return r.inner.FindByID(id)
}

func (r *failingMessageRepo) FindAllForReceiver(receiverID string, offset, limit int) ([]*entities.Message, error) {
	return r.inner.FindAllForReceiver(receiverID, offset, limit)
}

func (r *failingMessageRepo) FindAllFromSender(senderID string, offset, limit int) ([]*entities.Message, error) {
	return r.inner.FindAllFromSender(senderID, offset, limit)
}

func (r *failingMessageRepo) Update(m *entities.Message) error {
	return errors.New("injected messageRepo.Update failure")
}
