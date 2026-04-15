package repositories

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

// Idenpotency layer is used to ensure that retrying the same operation (e.g. due to network failure) won't cause duplicate side effects (e.g. double charge).
type PaymentIdempotencyRepository interface {
	Claim(operation string, key string) (bool, error)
	GetResponse(operation string, key string) (string, error)
	SetResponse(operation string, key string, response string) error
}

type PaymentSettlementRepository interface {
	Create(settlement *entities.PaymentSettlement) (*entities.PaymentSettlement, error)
	Update(settlement *entities.PaymentSettlement) (*entities.PaymentSettlement, error)
	FindByOperationAndKey(operation string, idempotencyKey string) (*entities.PaymentSettlement, error)
	FindByStatuses(statuses []entities.SettlementStatus, limit int) ([]entities.PaymentSettlement, error)
}
