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
	UserRepository               repositories.UserRepository
	GrindRepository              repositories.GrindRepository
	MessageRepository            repositories.MessageRepository
	PaymentMethodInfoRepository  repositories.PaymentMethodInfoRepository
	PaymentIdempotencyRepository repositories.PaymentIdempotencyRepository
	PaymentSettlementRepository  repositories.PaymentSettlementRepository
	ParticipationRepository      repositories.ParticipationRepository
}

func InitializeReposContainer(db *gorm.DB) error {
	if config.RepoType == "postgres" {
		Repos = &ReposContainer{
			UserRepository:               postgres.NewGormUserRepository(db),
			GrindRepository:              postgres.NewGormGrindRepository(db),
			MessageRepository:            postgres.NewGormMessageRepository(db),
			PaymentMethodInfoRepository:  postgres.NewGormStripePaymentInfoRepository(db),
			PaymentIdempotencyRepository: postgres.NewGormPaymentIdempotencyRepository(db),
			PaymentSettlementRepository:  postgres.NewGormPaymentSettlementRepository(db),
			ParticipationRepository:      postgres.NewGormParticipationRepository(db),
		}
		return nil
	}

	return errors.New("invalid repository type")
}
