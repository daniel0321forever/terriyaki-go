package repositories

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

type PaymentMethodInfoRepository interface {
	Create(paymentMethodInfo *entities.PaymentMethodInfo) (*entities.PaymentMethodInfo, error)
	FindByID(id string) (*entities.PaymentMethodInfo, error)
	FindByUserID(userID string) ([]entities.PaymentMethodInfo, error)
	Update(paymentMethodInfo *entities.PaymentMethodInfo) (*entities.PaymentMethodInfo, error)
	Delete(providerPaymentMethodID string) error
}
