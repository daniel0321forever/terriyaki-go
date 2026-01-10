package models

import (
	"errors"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/stripe/stripe-go/v84/paymentmethod"
	"gorm.io/gorm"
)

type StripePaymentInfo struct {
	gorm.Model
	UserID                string `json:"user_id" gorm:"not null"`
	StripeCustomerID      string `json:"stripe_customer_id" gorm:"not null"`
	StripePaymentMethodID string `json:"stripe_payment_method_id" gorm:"primaryKey;not null;unique"`
	Brand                 string `json:"brand" gorm:""`
	Last4                 string `json:"last4" gorm:""`
	ExpMonth              int    `json:"exp_month" gorm:""`
	ExpYear               int    `json:"exp_year" gorm:""`
}

type StripePaymentInfoRepository struct{}

func (*StripePaymentInfoRepository) Create(
	userID string,
	stripeCustomerID string,
	stripePaymentMethodID string,
) (*StripePaymentInfo, error) {
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

	stripePaymentInfo := StripePaymentInfo{
		UserID:                userID,
		StripeCustomerID:      stripeCustomerID,
		StripePaymentMethodID: stripePaymentMethodID,
		Brand:                 brand,
		Last4:                 last4,
		ExpMonth:              expMonth,
		ExpYear:               expYear,
	}

	result := database.Db.Create(&stripePaymentInfo)

	// check error
	if result.Error == nil {
		return nil, result.Error
	}

	return &stripePaymentInfo, nil
}

func (*StripePaymentInfoRepository) GetByID(id string) (*StripePaymentInfo, error) {
	var stripePaymentInfo StripePaymentInfo
	result := database.Db.Where("id = ?", id).First(&stripePaymentInfo)
	if result.Error != nil {
		return nil, result.Error
	}
	return &stripePaymentInfo, nil
}

func (*StripePaymentInfoRepository) GetByUserID(userID string) ([]StripePaymentInfo, error) {
	var stripePaymentInfos []StripePaymentInfo
	result := database.Db.Where("user_id = ?", userID).Find(&stripePaymentInfos)
	if result.Error != nil {
		return nil, result.Error
	}
	return stripePaymentInfos, nil
}

func (*StripePaymentInfoRepository) Update(stripePaymentInfo StripePaymentInfo) (*StripePaymentInfo, error) {
	result := database.Db.Model(&stripePaymentInfo).Where("payment_method_id = ?", stripePaymentInfo.StripePaymentMethodID).Updates(stripePaymentInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &stripePaymentInfo, nil
}

func (*StripePaymentInfoRepository) Delete(stripePaymentMethodID string) error {
	result := database.Db.Where("payment_method_id = ?", stripePaymentMethodID).Delete(&StripePaymentInfo{})
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
