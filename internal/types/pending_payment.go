package types

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

type PendingPayment struct {
	StripePaymentInfo entities.StripePaymentInfo
	PaymentAmount     int64
}
