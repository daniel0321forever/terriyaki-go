package services

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/customer"
	"github.com/stripe/stripe-go/v84/paymentintent"
	"github.com/stripe/stripe-go/v84/paymentmethod"
	"github.com/stripe/stripe-go/v84/setupintent"
	"github.com/stripe/stripe-go/v84/transfer"
)

// Compile-time check that Stripe adapter satisfies PaymentGatewayAdapter.
var _ PaymentGatewayAdapter = (*StripePaymentGatewayAdapter)(nil)

// "inherit" from PaymentGatewayAdapter
type StripePaymentGatewayAdapter struct {
	secretKey string
}

func NewStripePaymentGatewayAdapter(secretKey string) *StripePaymentGatewayAdapter {
	return &StripePaymentGatewayAdapter{secretKey: secretKey}
}

// `PaymentIntent`: object that tracks the entire lifecycle of a customer’s payment, from initiation to completion
func (a *StripePaymentGatewayAdapter) CreatePaymentIntent(amount int64) (string, error) {
	stripe.Key = a.secretKey
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", err
	}

	return pi.ClientSecret, nil
}

func (a *StripePaymentGatewayAdapter) CreateSaveCardIntent() (string, error) {
	stripe.Key = a.secretKey
	si, err := setupintent.New(&stripe.SetupIntentParams{Usage: stripe.String("off_session")})
	if err != nil {
		return "", err
	}

	return si.ClientSecret, nil
}

func (a *StripePaymentGatewayAdapter) CreateCustomer(name string, email string) (string, error) {
	stripe.Key = a.secretKey
	params := &stripe.CustomerParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}

	cus, err := customer.New(params)
	if err != nil {
		return "", err
	}

	return cus.ID, nil
}

func (a *StripePaymentGatewayAdapter) DescribePaymentMethod(paymentMethodID string) (*entities.PaymentMethodInfo, error) {
	stripe.Key = a.secretKey
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return nil, err
	}

	brand := ""
	last4 := ""
	expMonth := 0
	expYear := 0
	if pm.Card != nil {
		brand = string(pm.Card.Brand)
		last4 = string(pm.Card.Last4)
		expMonth = int(pm.Card.ExpMonth)
		expYear = int(pm.Card.ExpYear)
	}

	info := entities.NewPaymentMethodInfo(
		entities.PaymentProviderStripe,
		"",
		"",
		paymentMethodID,
		brand,
		last4,
		expMonth,
		expYear,
	)
	return info, nil
}

func (a *StripePaymentGatewayAdapter) AttachPaymentMethodToCustomer(paymentMethodID string, customerID string) error {
	stripe.Key = a.secretKey
	attachParams := &stripe.PaymentMethodAttachParams{Customer: stripe.String(customerID)}
	_, err := paymentmethod.Attach(paymentMethodID, attachParams)
	return err
}

func (a *StripePaymentGatewayAdapter) Charge(customerID string, paymentMethodID string, amount int64) (string, error) {
	stripe.Key = a.secretKey
	pi, err := paymentintent.New(&stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      stripe.String(customerID),
		PaymentMethod: stripe.String(paymentMethodID),
		OffSession:    stripe.Bool(true),
		Confirm:       stripe.Bool(true),
	})
	if err != nil {
		return "", err
	}

	return pi.ClientSecret, nil
}

func (a *StripePaymentGatewayAdapter) PayBack(destinationAccountID string, amount int64) error {
	stripe.Key = a.secretKey
	_, err := transfer.New(&stripe.TransferParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Destination: stripe.String(destinationAccountID),
	})

	return err
}
