package types

import "github.com/daniel0321forever/terriyaki-go/internal/models"

type PendingPayment struct {
	StripePaymentInfo models.StripePaymentInfo
	PaymentAmount     int64
}
