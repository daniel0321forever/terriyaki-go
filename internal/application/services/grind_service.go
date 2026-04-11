package services

import (
	"encoding/json"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"gorm.io/datatypes"
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

// Convert Grind entity to Grind DTO (including related entity fetching from DB)
func (s *GrindService) toGroupGrindDTO(grind *entities.Grind) (*dto.GroupGrindDTO, error) {
	participants, err := s.userRepo.FindByGrindID(grind.ID)
	if err != nil {
		return nil, config.ErrUserNotFound
	}

	return mappers.BuildGroupGrindDTO(grind, participants), nil
}

// Convert Grind entity to Participation DTO (including related entity fetching from DB)
func (s *GrindService) toParticipationDTO(participation *entities.Participation) *dto.ParticipationDTO {
	return mappers.BuildParticipationDTO(participation)
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

	// 3. Setup Creator and create all tasks for the creator (The person who made the API call)
	participation, err := entities.NewParticipation(request.CreatorID, grind.ID)
	_ = s.participationRepo.Create(participation)

	// assign participants
	creator, err := s.userRepo.FindById(request.CreatorID)
	if err != nil {
		return nil, err
	}
	grind.Participants = []entities.User{*creator}

	// create tasks
	var tasks []entities.Task = make([]entities.Task, 0, request.Duration)
	for i := 0; i < request.Duration; i++ {
		task, err := entities.NewTask(request.CreatorID, grind.ID, request.StartDate.AddDate(0, 0, i))
		if err != nil {
			return nil, err
		}

		// assign leetcode problem to the task
		problem, err := utils.GetRandomProblemFromList("neetcode250")
		if err != nil {
			return nil, err
		}
		task.ProblemTitle = &problem.Title
		problemURL := "https://leetcode.com/problems/" + problem.Slug
		task.ProblemURL = &problemURL
		task.ProblemDifficulty = &problem.Difficulty
		topicTagsJSON, _ := json.Marshal(problem.TopicTags)
		task.ProblemTopicTags = datatypes.JSON(topicTagsJSON)

		s.taskRepo.Create(task)
		tasks = append(tasks, *task)
	}
	grind.Tasks = tasks

	return s.toGroupGrindDTO(grind)
}

// GetOngoingGrindByUserID combines data fetching with business rules
func (s *GrindService) GetOngoingGrindByUserID(request dto.GetOngoingGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindLatestByUserID(request.UserID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}

	// Logic: Check if the grind has ended (Business Rule)
	endDate := grind.StartDate.AddDate(0, 0, int(grind.Duration))
	if endDate.Before(time.Now().UTC()) {
		return nil, config.ErrNoOngoingGrind
	}

	// Logic: Check if user quitted
	record, err := s.participationRepo.FindByUserAndGrind(request.UserID, grind.ID)
	if err != nil || record.Quitted {
		return nil, config.ErrUserNotParticipatingOrQuit
	}

	// user's tasks
	tasks, err := s.taskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
	if err != nil {
		return nil, config.ErrTasksNotFound
	}
	grind.Tasks = tasks

	return s.toGroupGrindDTO(grind)
}

/*
 * Get a grind by ID and user ID
 * @param request dto.GetGrindDTO
 * @return *dto.GroupGrindDTO, error
 * @throws ErrGrindNotFound
 * @throws ErrTasksNotFound
 */
func (s *GrindService) GetGrind(request dto.GetGrindDTO) (*dto.GroupGrindDTO, error) {
	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}

	// user's tasks
	tasks, err := s.taskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
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
		tasks, err := s.taskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
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
	err = s.grindRepo.Update(grind)
	if err != nil {
		return nil, config.ErrGrindUpdateFailed
	}
	tasks, err := s.taskRepo.FindByGrindIDAndParticipantID(grind.ID, request.UserID)
	if err != nil {
		return nil, config.ErrTasksNotFound
	}
	grind.Tasks = tasks
	return s.toGroupGrindDTO(grind)
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
		return config.ErrParticipationAlreadyExists(request.UserID, request.GrindID)
	}

	// 2. Make sure user and grind exists
	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return config.ErrUserNotFound
	}

	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return config.ErrGrindNotFound
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

func (s *GrindService) QuitGrind(request dto.QuitGrindDTO) (*dto.ParticipationDTO, error) {
	// find the grind
	grind, err := s.grindRepo.FindById(request.GrindID)
	if err != nil {
		return nil, config.ErrGrindNotFound
	}

	// 1. Check if the user is a participant of the grind
	participation, err := s.participationRepo.FindByUserAndGrind(request.UserID, request.GrindID)
	if err != nil {
		return nil, config.ErrParticipationNotFound
	}
	if participation == nil {
		return nil, config.ErrUserIsNotParticipant
	}

	// 2. Update the participation record
	participation.Quitted = true
	participation.QuittedAt = time.Now()
	participation.TotalPenalty = int(grind.Budget)
	err = s.participationRepo.Update(participation)
	if err != nil {
		return nil, config.ErrParticipationUpdateFailed
	}

	return s.toParticipationDTO(participation), nil
}
