package repositories

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

type StripePaymentInfoRepository interface {
	Create(userID string, stripeCustomerID string, stripePaymentMethodID string) (*entities.StripePaymentInfo, error)
	FindByID(id string) (*entities.StripePaymentInfo, error)
	FindByUserID(userID string) ([]entities.StripePaymentInfo, error)
	Update(stripePaymentInfo *entities.StripePaymentInfo) (*entities.StripePaymentInfo, error)
	Delete(stripePaymentMethodID string) error
}
