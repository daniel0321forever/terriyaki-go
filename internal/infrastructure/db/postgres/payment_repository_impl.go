package postgres

import (
	"errors"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

// PaymentMethodInfoSchema is the canonical provider-neutral persistence shape.
type PaymentMethodInfoSchema struct {
	gorm.Model
	UserID                  string `json:"user_id" gorm:"not null"`
	Provider                string `json:"provider" gorm:"not null"`
	ProviderCustomerID      string `json:"provider_customer_id" gorm:""`
	ProviderPaymentMethodID string `json:"provider_payment_method_id" gorm:"not null;uniqueIndex"`
	MethodType              string `json:"method_type" gorm:""`
	Brand                   string `json:"brand" gorm:""`
	Last4                   string `json:"last4" gorm:""`
	ExpMonth                int    `json:"exp_month" gorm:""`
	ExpYear                 int    `json:"exp_year" gorm:""`
	Network                 string `json:"network" gorm:""`
	WalletAddress           string `json:"wallet_address" gorm:""`
}

func (PaymentMethodInfoSchema) TableName() string { return "payment_method_infos" }

type GormStripePaymentInfoRepository struct {
	db *gorm.DB
}

func NewGormStripePaymentInfoRepository(db *gorm.DB) *GormStripePaymentInfoRepository {
	return &GormStripePaymentInfoRepository{db: db}
}
func (*GormStripePaymentInfoRepository) Create(paymentMethodInfo *entities.PaymentMethodInfo) (*entities.PaymentMethodInfo, error) {
	paymentMethodInfoSchema := PaymentMethodInfoSchema{
		UserID:                  paymentMethodInfo.UserID,
		Provider:                string(paymentMethodInfo.Provider),
		ProviderCustomerID:      paymentMethodInfo.ProviderCustomerID,
		ProviderPaymentMethodID: paymentMethodInfo.ProviderPaymentMethodID,
		MethodType:              paymentMethodInfo.MethodType,
		Brand:                   paymentMethodInfo.Brand,
		Last4:                   paymentMethodInfo.Last4,
		ExpMonth:                paymentMethodInfo.ExpMonth,
		ExpYear:                 paymentMethodInfo.ExpYear,
		Network:                 paymentMethodInfo.Network,
		WalletAddress:           paymentMethodInfo.WalletAddress,
	}

	result := Db.Create(&paymentMethodInfoSchema)

	if result.Error != nil {
		return nil, result.Error
	}

	return paymentMethodInfo, nil
}

func (*GormStripePaymentInfoRepository) FindByID(id string) (*entities.PaymentMethodInfo, error) {
	var paymentMethodInfoSchema PaymentMethodInfoSchema
	result := Db.Where("id = ?", id).First(&paymentMethodInfoSchema)
	if result.Error != nil {
		return nil, errors.New("payment method info not found")
	}
	paymentInfo := entities.NewPaymentMethodInfo(
		entities.PaymentProvider(paymentMethodInfoSchema.Provider),
		paymentMethodInfoSchema.MethodType,
		paymentMethodInfoSchema.UserID,
		paymentMethodInfoSchema.ProviderCustomerID,
		paymentMethodInfoSchema.ProviderPaymentMethodID,
		paymentMethodInfoSchema.Brand,
		paymentMethodInfoSchema.Last4,
		paymentMethodInfoSchema.ExpMonth,
		paymentMethodInfoSchema.ExpYear,
	)
	paymentInfo.WalletAddress = paymentMethodInfoSchema.WalletAddress
	return paymentInfo, nil
}

func (*GormStripePaymentInfoRepository) FindByUserID(userID string) ([]entities.PaymentMethodInfo, error) {
	var paymentMethodInfoSchemas []PaymentMethodInfoSchema
	result := Db.Where("user_id = ?", userID).Find(&paymentMethodInfoSchemas)
	if result.Error != nil {
		return nil, errors.New("payment method infos not found")
	}

	paymentInfos := make([]entities.PaymentMethodInfo, 0, len(paymentMethodInfoSchemas))
	for _, info := range paymentMethodInfoSchemas {
		paymentInfo := entities.NewPaymentMethodInfo(
			entities.PaymentProvider(info.Provider),
			info.MethodType,
			info.UserID,
			info.ProviderCustomerID,
			info.ProviderPaymentMethodID,
			info.Brand,
			info.Last4,
			info.ExpMonth,
			info.ExpYear,
		)
		paymentInfo.Network = info.Network
		paymentInfo.WalletAddress = info.WalletAddress
		paymentInfos = append(paymentInfos, *paymentInfo)
	}

	return paymentInfos, nil
}

func (*GormStripePaymentInfoRepository) Update(paymentMethodInfo *entities.PaymentMethodInfo) (*entities.PaymentMethodInfo, error) {
	paymentMethodInfoSchema := PaymentMethodInfoSchema{
		UserID:                  paymentMethodInfo.UserID,
		Provider:                string(paymentMethodInfo.Provider),
		ProviderCustomerID:      paymentMethodInfo.ProviderCustomerID,
		ProviderPaymentMethodID: paymentMethodInfo.ProviderPaymentMethodID,
		MethodType:              paymentMethodInfo.MethodType,
		Brand:                   paymentMethodInfo.Brand,
		Last4:                   paymentMethodInfo.Last4,
		ExpMonth:                paymentMethodInfo.ExpMonth,
		ExpYear:                 paymentMethodInfo.ExpYear,
		Network:                 paymentMethodInfo.Network,
		WalletAddress:           paymentMethodInfo.WalletAddress,
	}

	result := Db.Model(&paymentMethodInfoSchema).Where("provider_payment_method_id = ?", paymentMethodInfo.ProviderPaymentMethodID).Updates(&paymentMethodInfoSchema)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return paymentMethodInfo, nil
}

func (*GormStripePaymentInfoRepository) Delete(stripePaymentMethodID string) error {
	result := Db.Where("provider_payment_method_id = ?", stripePaymentMethodID).Delete(&PaymentMethodInfoSchema{})
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
