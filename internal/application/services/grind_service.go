package services

import (
	"errors"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
)

type GrindService struct {
	grindRepo         repositories.GrindRepository
	userRepo          repositories.UserRepository
	taskRepo          repositories.TaskRepository
	participationRepo repositories.ParticipationRepository
	messageRepo       repositories.MessageRepository
}

func NewGrindService(
	grindRepo repositories.GrindRepository,
	userRepo repositories.UserRepository,
	taskRepo repositories.TaskRepository,
	participationRepo repositories.ParticipationRepository,
	messageRepo repositories.MessageRepository,
) *GrindService {
	return &GrindService{
		grindRepo:         grindRepo,
		userRepo:          userRepo,
		taskRepo:          taskRepo,
		participationRepo: participationRepo,
		messageRepo:       messageRepo,
	}
}

// Create grind with just one single user (initialization)
func (s *GrindService) CreateGroupGrind(request dto.CreateGrindDTO) (*dto.GroupGrindDTO, error) {
	// 1. Create grind
	grind, err := entities.NewGrind(request.Duration, request.Budget, request.StartDate)
	if err != nil {
		return nil, err
	}

	// 2. Infrastructure: Save the Grind to db
	err = s.grindRepo.Create(grind)
	if err != nil {
		return nil, err
	}

	// 3. Setup Creator (The person who made the API call)
	participation, err := entities.NewParticipation(request.CreatorID, grind.ID)
	_ = s.participationRepo.Create(participation)
	for i := 0; i < request.Duration; i++ {
		task, err := entities.NewTask(request.CreatorID, grind.ID, request.StartDate.AddDate(0, 0, i))
		if err != nil {
			return nil, err
		}
		s.taskRepo.Create(task)
	}

	return mappers.GrindToGroupGrindDTO(grind), nil
}

// GetOngoingGrindByUserID combines data fetching with business rules
func (s *GrindService) GetOngoingGrindByUserID(request dto.GetOngoingGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindLatestByUserID(request.UserID)
	if err != nil {
		return nil, errors.New("grind not found")
	}

	// Logic: Check if the grind has ended (Business Rule)
	endDate := grind.StartDate.AddDate(0, 0, int(grind.Duration))
	if endDate.Before(time.Now().UTC()) {
		return nil, errors.New("no ongoing grind found")
	}

	// Logic: Check if user quitted
	record, err := s.participationRepo.FindByUserAndGrind(request.UserID, grind.ID)
	if err != nil || record.Quitted {
		return nil, errors.New("user not participating or quitted")
	}

	return mappers.GrindToGroupGrindDTO(grind), nil
}

func (s *GrindService) GetGrind(request dto.GetGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return nil, errors.New("grind not found")
	}
	return mappers.GrindToGroupGrindDTO(grind), nil
}

func (s *GrindService) GetAllUserGrinds(request dto.GetAllUserGrindsDTO) ([]*dto.GroupGrindDTO, error) {
	grinds, err := s.grindRepo.FindAllByUserID(request.UserID)
	if err != nil {
		return nil, errors.New("grind not found")
	}
	var output []*dto.GroupGrindDTO
	for _, grind := range grinds {
		output = append(output, mappers.GrindToGroupGrindDTO(grind))
	}
	return output, nil
}

func (s *GrindService) UpdateGrind(request dto.UpdateGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return nil, errors.New("grind not found")
	}
	grind.Duration = int32(request.Duration)
	grind.Budget = int32(request.Budget)
	err = s.grindRepo.Update(grind)
	if err != nil {
		return nil, errors.New("grind update failed")
	}
	return mappers.GrindToGroupGrindDTO(grind), nil
}

func (s *GrindService) DeleteGrind(request dto.DeleteGrindDTO) error {
	// 1. Delete associated tasks first
	err := s.taskRepo.DeleteByGrindID(request.GrindID)
	if err != nil {
		return err
	}

	// 2. Clear many-to-many associations (Records)
	err = s.participationRepo.DeleteByGrindID(request.GrindID)
	if err != nil {
		return err
	}

	// 3. Delete the actual grind
	err = s.grindRepo.Delete(request.GrindID)
	return err
}

func (s *GrindService) DeleteAllGrinds() error {
	return s.grindRepo.DeleteAll()
}

// AddParticipation handles the side-effects of a new joiner
func (s *GrindService) AddParticipation(request dto.AddParticipationDTO) error {
	// 1. Check existing
	participation, _ := s.participationRepo.FindByUserAndGrind(request.UserID, request.GrindID)
	if participation != nil {
		return errors.New("already exists participation record for " + request.UserID + " and " + request.GrindID)
	}

	// 2. Make sure user and grind exists
	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return errors.New("grind not found")
	}

	// 3. Save Record
	participation, err = entities.NewParticipation(user.ID, grind.ID)
	if err != nil {
		return err
	}
	err = s.participationRepo.Create(participation)
	if err != nil {
		return err
	}

	// 4. Create tasks for the new participant
	for i := 0; i < int(grind.Duration); i++ {
		taskDate := grind.StartDate.AddDate(0, 0, i)
		task, err := entities.NewTask(user.ID, grind.ID, taskDate)
		if err != nil {
			return err
		}
		err = s.taskRepo.Create(task)
		if err != nil {
			return err
		}
	}

	return nil
}
