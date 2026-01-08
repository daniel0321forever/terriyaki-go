package types

import "time"

type PaymentInfo struct {
	CustomerID      string
	PaymentIntentID string
	DueDate         time.Time
}
