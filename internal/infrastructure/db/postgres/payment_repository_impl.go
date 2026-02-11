package postgres

import (
	"errors"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stripe/stripe-go/v84/paymentmethod"
	"gorm.io/gorm"
)

type StripePaymentInfoSchema struct {
	gorm.Model
	UserID                string `json:"user_id" gorm:"not null"`
	StripeCustomerID      string `json:"stripe_customer_id" gorm:"not null"`
	StripePaymentMethodID string `json:"stripe_payment_method_id" gorm:"primaryKey;not null;unique"`
	Brand                 string `json:"brand" gorm:""`
	Last4                 string `json:"last4" gorm:""`
	ExpMonth              int    `json:"exp_month" gorm:""`
	ExpYear               int    `json:"exp_year" gorm:""`
}

func (StripePaymentInfoSchema) TableName() string { return "stripe_payment_info" }

type GormStripePaymentInfoRepository struct {
	db *gorm.DB
}

func NewGormStripePaymentInfoRepository(db *gorm.DB) *GormStripePaymentInfoRepository {
	return &GormStripePaymentInfoRepository{db: db}
}
func (*GormStripePaymentInfoRepository) Create(
	userID string,
	stripeCustomerID string,
	stripePaymentMethodID string,
) (*entities.StripePaymentInfo, error) {
	// 1. Fetch the payment method object from Stripe
	pm, err := paymentmethod.Get(stripePaymentMethodID, nil)
	if err != nil {
		// Handle error
		return nil, err
	}

	// 2. Access the fields directly from the 'Card' property
	brand := string(pm.Card.Brand) // e.g., "visa"
	last4 := string(pm.Card.Last4) // e.g., "4242"
	expMonth := int(pm.Card.ExpMonth)
	expYear := int(pm.Card.ExpYear)

	stripePaymentInfo := StripePaymentInfoSchema{
		UserID:                userID,
		StripeCustomerID:      stripeCustomerID,
		StripePaymentMethodID: stripePaymentMethodID,
		Brand:                 brand,
		Last4:                 last4,
		ExpMonth:              expMonth,
		ExpYear:               expYear,
	}

	result := Db.Create(&stripePaymentInfo)

	// check error
	if result.Error == nil {
		return nil, result.Error
	}

	return &entities.StripePaymentInfo{
		UserID:                userID,
		StripeCustomerID:      stripeCustomerID,
		StripePaymentMethodID: stripePaymentMethodID,
		Brand:                 brand,
		Last4:                 last4,
		ExpMonth:              expMonth,
		ExpYear:               expYear,
	}, nil
}

func (*GormStripePaymentInfoRepository) FindByID(id string) (*entities.StripePaymentInfo, error) {
	var stripePaymentInfo entities.StripePaymentInfo
	result := Db.Where("id = ?", id).First(&stripePaymentInfo)
	if result.Error != nil {
		return nil, errors.New("stripe payment info not found")
	}
	return &stripePaymentInfo, nil
}

func (*GormStripePaymentInfoRepository) FindByUserID(userID string) ([]entities.StripePaymentInfo, error) {
	var stripePaymentInfos []entities.StripePaymentInfo
	result := Db.Where("user_id = ?", userID).Find(&stripePaymentInfos)
	if result.Error != nil {
		return nil, errors.New("stripe payment infos not found")
	}
	return stripePaymentInfos, nil
}

func (*GormStripePaymentInfoRepository) Update(stripePaymentInfo *entities.StripePaymentInfo) (*entities.StripePaymentInfo, error) {
	result := Db.Model(&stripePaymentInfo).Where("payment_method_id = ?", stripePaymentInfo.StripePaymentMethodID).Updates(stripePaymentInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return stripePaymentInfo, nil
}

func (*GormStripePaymentInfoRepository) Delete(stripePaymentMethodID string) error {
	result := Db.Where("payment_method_id = ?", stripePaymentMethodID).Delete(&StripePaymentInfoSchema{})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
