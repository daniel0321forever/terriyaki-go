package services

import (
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"gorm.io/gorm"
)

type GrindService struct {
	db                *gorm.DB
	grindRepo         repositories.GrindRepository
	userRepo          repositories.UserRepository
	habitTaskRepo     repositories.HabitTaskRepository
	participationRepo repositories.ParticipationRepository
	messageRepo       repositories.MessageRepository
}

func NewGrindService(
	db *gorm.DB,
	grindRepo repositories.GrindRepository,
	userRepo repositories.UserRepository,
	habitTaskRepo repositories.HabitTaskRepository,
	participationRepo repositories.ParticipationRepository,
	messageRepo repositories.MessageRepository,
) *GrindService {
	return &GrindService{
		db:                db,
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

	var result *dto.GroupGrindDTO

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		grindRepo := getGrindRepo(s.grindRepo, tx)
		partRepo := getParticipationRepo(s.participationRepo, tx)
		habitTaskRepo := getHabitTaskRepo(s.habitTaskRepo, tx)

		if err := grindRepo.Create(grind); err != nil {
			return err
		}

		participation, err := entities.NewParticipation(request.CreatorID, grind.ID)
		if err != nil {
			return err
		}
		if err := partRepo.Create(participation); err != nil {
			return err
		}

		tasks := make([]entities.HabitTask, 0, request.Duration)
		for i := 0; i < request.Duration; i++ {
			task, err := entities.NewHabitTask(request.CreatorID, grind.ID, request.StartDate.AddDate(0, 0, i))
			if err != nil {
				return err
			}
			if err := habitTaskRepo.Create(task); err != nil {
				return err
			}
			tasks = append(tasks, *task)
		}
		grind.Tasks = tasks

		creator, err := s.userRepo.FindById(request.CreatorID)
		if err != nil {
			return err
		}
		grind.Participants = []entities.User{*creator}

		dto, dtoErr := s.toGroupGrindDTO(grind)
		if dtoErr != nil {
			return dtoErr
		}
		result = dto
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return result, nil
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
	_, err = s.participationRepo.FindByUserAndGrind(request.UserID, request.GrindID)
	if err != nil {
		return nil, config.ErrUserIsNotParticipant
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
	return s.db.Transaction(func(tx *gorm.DB) error {
		habitTaskRepo := getHabitTaskRepo(s.habitTaskRepo, tx)
		partRepo := getParticipationRepo(s.participationRepo, tx)
		grindRepo := getGrindRepo(s.grindRepo, tx)

		if err := habitTaskRepo.DeleteByGrindID(request.GrindID); err != nil {
			return err
		}
		if err := partRepo.DeleteByGrindID(request.GrindID); err != nil {
			return err
		}
		return grindRepo.Delete(request.GrindID)
	})
}

func (s *GrindService) DeleteAllGrinds() error {
	return s.grindRepo.DeleteAll()
}

func (s *GrindService) AddParticipation(request dto.AddParticipationDTO) error {
	// Guard reads outside the transaction
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

	// Wrap only the writes in a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		partRepo := getParticipationRepo(s.participationRepo, tx)
		habitTaskRepo := getHabitTaskRepo(s.habitTaskRepo, tx)

		participation, err := entities.NewParticipation(user.ID, grind.ID)
		if err != nil {
			return err
		}
		if err := partRepo.Create(participation); err != nil {
			return err
		}

		for i := 0; i < int(grind.Duration); i++ {
			task, err := entities.NewHabitTask(user.ID, grind.ID, grind.StartDate.AddDate(0, 0, i))
			if err != nil {
				return err
			}
			if err := habitTaskRepo.Create(task); err != nil {
				return err
			}
		}
		return nil
	})
}

// AcceptInvitation orchestrates the acceptance of a grind invitation atomically.
// It creates participation + habit tasks + updates message status + creates accepted notification
// all within a single DB transaction.
func (s *GrindService) AcceptInvitation(
	addParticipationReq dto.AddParticipationDTO,
	updateMsgReq dto.UpdateMessageInvitationAcceptedStatusDTO,
	createAcceptedMsgReq dto.CreateInvitationAcceptedMessageDTO,
	messageRepo repositories.MessageRepository,
) error {
	// Guard reads outside the transaction
	existing, _ := s.participationRepo.FindByUserAndGrind(addParticipationReq.UserID, addParticipationReq.GrindID)
	if existing != nil {
		return config.ErrParticipationAlreadyExists(addParticipationReq.UserID, addParticipationReq.GrindID)
	}

	user, err := s.userRepo.FindById(addParticipationReq.UserID)
	if err != nil {
		return config.ErrUserNotFound
	}

	grind, err := s.grindRepo.FindById(addParticipationReq.GrindID)
	if err != nil {
		return config.ErrGrindNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		partRepo := getParticipationRepo(s.participationRepo, tx)
		habitTaskRepo := getHabitTaskRepo(s.habitTaskRepo, tx)
		msgRepo := getMessageRepo(messageRepo, tx)

		// Create participation
		participation, err := entities.NewParticipation(user.ID, grind.ID)
		if err != nil {
			return err
		}
		if err := partRepo.Create(participation); err != nil {
			return err
		}

		// Create habit tasks for each day of the grind
		for i := 0; i < int(grind.Duration); i++ {
			task, err := entities.NewHabitTask(user.ID, grind.ID, grind.StartDate.AddDate(0, 0, i))
			if err != nil {
				return err
			}
			if err := habitTaskRepo.Create(task); err != nil {
				return err
			}
		}

		// Update original invitation message status to accepted
		inviteMsg, err := msgRepo.FindByID(updateMsgReq.MessageID)
		if err != nil {
			return err
		}
		inviteMsg.InvitationAccepted = updateMsgReq.Accepted
		if err := msgRepo.Update(inviteMsg); err != nil {
			return err
		}

		// Create accepted notification message to invitor
		acceptedMsg, err := entities.NewMessage(
			createAcceptedMsgReq.AccepterID,
			createAcceptedMsgReq.InvitorID,
			createAcceptedMsgReq.AccepterID+" accepted your invitation",
			"invitation_accepted",
			createAcceptedMsgReq.GrindID,
			true,
			false,
		)
		if err != nil {
			return err
		}
		return msgRepo.Create(acceptedMsg)
	})
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

// MessageRepo returns the message repository for use in controller-level orchestration.
func (s *GrindService) MessageRepo() repositories.MessageRepository {
	return s.messageRepo
}

func getGrindRepo(r repositories.GrindRepository, tx *gorm.DB) repositories.GrindRepository {
	if txRepo, ok := r.(interface {
		WithTx(tx *gorm.DB) repositories.GrindRepository
	}); ok {
		return txRepo.WithTx(tx)
	}
	return r
}

func getParticipationRepo(r repositories.ParticipationRepository, tx *gorm.DB) repositories.ParticipationRepository {
	if txRepo, ok := r.(interface {
		WithTx(tx *gorm.DB) repositories.ParticipationRepository
	}); ok {
		return txRepo.WithTx(tx)
	}
	return r
}

func getHabitTaskRepo(r repositories.HabitTaskRepository, tx *gorm.DB) repositories.HabitTaskRepository {
	if txRepo, ok := r.(interface {
		WithTx(tx *gorm.DB) repositories.HabitTaskRepository
	}); ok {
		return txRepo.WithTx(tx)
	}
	return r
}

func getMessageRepo(r repositories.MessageRepository, tx *gorm.DB) repositories.MessageRepository {
	if txRepo, ok := r.(interface {
		WithTx(tx *gorm.DB) repositories.MessageRepository
	}); ok {
		return txRepo.WithTx(tx)
	}
	return r
}
