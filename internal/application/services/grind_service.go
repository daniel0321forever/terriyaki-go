package services

import (
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
)

type GrindService struct {
	grindRepo         repositories.GrindRepository
	userRepo          repositories.UserRepository
	habitTaskRepo     repositories.HabitTaskRepository
	participationRepo repositories.ParticipationRepository
	messageRepo       repositories.MessageRepository
}

func NewGrindService(
	grindRepo repositories.GrindRepository,
	userRepo repositories.UserRepository,
	habitTaskRepo repositories.HabitTaskRepository,
	participationRepo repositories.ParticipationRepository,
	messageRepo repositories.MessageRepository,
) *GrindService {
	return &GrindService{
		grindRepo:         grindRepo,
		userRepo:          userRepo,
		habitTaskRepo:     habitTaskRepo,
		participationRepo: participationRepo,
		messageRepo:       messageRepo,
	}
}

func (s *GrindService) toGroupGrindDTO(grind *entities.Grind) (*dto.GroupGrindDTO, error) {
	participants, err := s.userRepo.FindByGrindID(grind.ID)
	if err != nil {
		return nil, config.ErrUserNotFound
	}
	return mappers.BuildGroupGrindDTO(grind, participants), nil
}

func (s *GrindService) toParticipationDTO(participation *entities.Participation) *dto.ParticipationDTO {
	return mappers.BuildParticipationDTO(participation)
}

func (s *GrindService) CreateGroupGrind(request dto.CreateGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := entities.NewGrind(request.Duration, request.Budget, request.StartDate)
	if err != nil {
		return nil, err
	}
	if err := s.grindRepo.Create(grind); err != nil {
		return nil, err
	}

	participation, err := entities.NewParticipation(request.CreatorID, grind.ID)
	if err != nil {
		return nil, err
	}
	_ = s.participationRepo.Create(participation)

	creator, err := s.userRepo.FindById(request.CreatorID)
	if err != nil {
		return nil, err
	}
	grind.Participants = []entities.User{*creator}

	tasks := make([]entities.HabitTask, 0, request.Duration)
	for i := 0; i < request.Duration; i++ {
		task, err := entities.NewHabitTask(request.CreatorID, grind.ID, request.StartDate.AddDate(0, 0, i))
		if err != nil {
			return nil, err
		}
		if err := s.habitTaskRepo.Create(task); err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	grind.Tasks = tasks

	return s.toGroupGrindDTO(grind)
}

func (s *GrindService) GetOngoingGrindByUserID(request dto.GetOngoingGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindLatestByUserID(request.UserID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}

	endDate := grind.StartDate.AddDate(0, 0, int(grind.Duration))
	if endDate.Before(time.Now().UTC()) {
		return nil, config.ErrNoOngoingGrind
	}

	record, err := s.participationRepo.FindByUserAndGrind(request.UserID, grind.ID)
	if err != nil || record.Quitted {
		return nil, config.ErrUserNotParticipatingOrQuit
	}

	tasks, err := s.habitTaskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
	if err != nil {
		return nil, config.ErrTasksNotFound
	}
	grind.Tasks = tasks

	return s.toGroupGrindDTO(grind)
}

func (s *GrindService) GetGrind(request dto.GetGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}
	tasks, err := s.habitTaskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
	if err != nil {
		return nil, config.ErrTasksNotFound
	}
	grind.Tasks = tasks
	return s.toGroupGrindDTO(grind)
}

func (s *GrindService) GetAllUserGrinds(request dto.GetAllUserGrindsDTO) (map[string]*dto.GroupGrindDTO, error) {
	grinds, err := s.grindRepo.FindAllByUserID(request.UserID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}
	output := make(map[string]*dto.GroupGrindDTO)
	for _, grind := range grinds {
		tasks, err := s.habitTaskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
		if err != nil {
			return nil, config.ErrTasksNotFound
		}
		grind.Tasks = tasks
		grindDTO, dtoErr := s.toGroupGrindDTO(grind)
		if dtoErr != nil {
			return nil, dtoErr
		}
		output[grind.ID] = grindDTO
	}
	return output, nil
}

func (s *GrindService) UpdateGrind(request dto.UpdateGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}
	grind.Duration = int32(request.Duration)
	grind.Budget = int32(request.Budget)
	if err := s.grindRepo.Update(grind); err != nil {
		return nil, config.ErrGrindUpdateFailed
	}
	tasks, err := s.habitTaskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
	if err != nil {
		return nil, config.ErrTasksNotFound
	}
	grind.Tasks = tasks
	return s.toGroupGrindDTO(grind)
}

func (s *GrindService) DeleteGrind(request dto.DeleteGrindDTO) error {
	if err := s.habitTaskRepo.DeleteByGrindID(request.GrindID); err != nil {
		return err
	}
	if err := s.participationRepo.DeleteByGrindID(request.GrindID); err != nil {
		return err
	}
	return s.grindRepo.Delete(request.GrindID)
}

func (s *GrindService) DeleteAllGrinds() error {
	return s.grindRepo.DeleteAll()
}

func (s *GrindService) AddParticipation(request dto.AddParticipationDTO) error {
	existing, _ := s.participationRepo.FindByUserAndGrind(request.UserID, request.GrindID)
	if existing != nil {
		return config.ErrParticipationAlreadyExists(request.UserID, request.GrindID)
	}

	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return config.ErrUserNotFound
	}

	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return config.ErrGrindNotFound
	}

	participation, err := entities.NewParticipation(user.ID, grind.ID)
	if err != nil {
		return err
	}
	if err := s.participationRepo.Create(participation); err != nil {
		return err
	}

	for i := 0; i < int(grind.Duration); i++ {
		task, err := entities.NewHabitTask(user.ID, grind.ID, grind.StartDate.AddDate(0, 0, i))
		if err != nil {
			return err
		}
		if err := s.habitTaskRepo.Create(task); err != nil {
			return err
		}
	}
	return nil
}

func (s *GrindService) QuitGrind(request dto.QuitGrindDTO) (*dto.ParticipationDTO, error) {
	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}

	participation, err := s.participationRepo.FindByUserAndGrind(request.UserID, request.GrindID)
	if err != nil {
		return nil, config.ErrParticipationNotFound
	}
	if participation == nil {
		return nil, config.ErrUserIsNotParticipant
	}

	participation.Quitted = true
	participation.QuittedAt = time.Now()
	participation.TotalPenalty = int(grind.Budget)
	if err := s.participationRepo.Update(participation); err != nil {
		return nil, config.ErrParticipationUpdateFailed
	}

	return s.toParticipationDTO(participation), nil
}
