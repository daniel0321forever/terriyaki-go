package services

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

// PaymentGatewayAdapter abstracts provider-specific payment operations.
type PaymentGatewayAdapter interface {
	CreatePaymentIntent(amount int64) (string, error)
	CreateSaveCardIntent() (string, error)
	CreateCustomer(name string, email string) (string, error)
	DescribePaymentMethod(paymentMethodID string) (*entities.PaymentMethodInfo, error)
	AttachPaymentMethodToCustomer(paymentMethodID string, customerID string) error
	Charge(customerID string, paymentMethodID string, amount int64) (string, error)
	PayBack(destinationAccountID string, amount int64) error
}
