package entities

type StripePaymentInfo struct {
	UserID                string `json:"user_id" gorm:"not null"`
	StripeCustomerID      string `json:"stripe_customer_id" gorm:"not null"`
	StripePaymentMethodID string `json:"stripe_payment_method_id" gorm:"primaryKey;not null;unique"`
	Brand                 string `json:"brand" gorm:""`
	Last4                 string `json:"last4" gorm:""`
	ExpMonth              int    `json:"exp_month" gorm:""`
	ExpYear               int    `json:"exp_year" gorm:""`
}

func NewStripePaymentInfo(userID string, stripeCustomerID string, stripePaymentMethodID string, brand string, last4 string, expMonth int, expYear int) *StripePaymentInfo {
	return &StripePaymentInfo{
		UserID:                userID,
		StripeCustomerID:      stripeCustomerID,
		StripePaymentMethodID: stripePaymentMethodID,
		Brand:                 brand,
		Last4:                 last4,
		ExpMonth:              expMonth,
		ExpYear:               expYear,
	}
}
