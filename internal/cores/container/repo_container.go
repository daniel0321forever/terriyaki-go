package container

import (
	"errors"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"gorm.io/gorm"
)

var Repos *ReposContainer = nil

type ReposContainer struct {
	UserRepository              repositories.UserRepository
	GrindRepository             repositories.GrindRepository
	TaskRepository              repositories.TaskRepository
	MessageRepository           repositories.MessageRepository
	InterviewSessionRepository  repositories.InterviewSessionRepository
	StripePaymentInfoRepository repositories.StripePaymentInfoRepository
	ParticipationRepository     repositories.ParticipationRepository
}

func InitializeReposContainer(db *gorm.DB) error {
	if config.RepoType == "postgres" {
		Repos = &ReposContainer{
			UserRepository:              postgres.NewGormUserRepository(db),
			GrindRepository:             postgres.NewGormGrindRepository(db),
			TaskRepository:              postgres.NewGormTaskRepository(db),
			MessageRepository:           postgres.NewGormMessageRepository(db),
			InterviewSessionRepository:  postgres.NewGormInterviewSessionRepository(db),
			StripePaymentInfoRepository: postgres.NewGormStripePaymentInfoRepository(db),
			ParticipationRepository:     postgres.NewGormParticipationRepository(db),
		}
		return nil
	}

	return errors.New("invalid repository type")
}
