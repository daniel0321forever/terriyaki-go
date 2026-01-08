package services

import (
	"github.com/daniel0321forever/terriyaki-go/internal/types"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/paymentintent"
	"github.com/stripe/stripe-go/v84/setupintent"
	"github.com/stripe/stripe-go/v84/transfer"
)

type IPaymentService interface {
	/*
		Create a payment intent
		@param amount - the amount to create a payment intent
		@return the payment intent id
	*/
	CreatePaymentIntent(amount int64) (string, error)

	/*
		Create a save card intent
		@return the save card intent id
	*/
	CreateSaveCardIntent() (string, error)

	/*
		Save a card
		@param paymentInfo - the payment info
		@return the card id
	*/
	SaveCard(paymentInfo types.PaymentInfo) error

	/*
		Charge a payment intent
		@param amount - the amount to charge
		@return the payment intent id
	*/
	Charge(paymentInfo types.PaymentInfo, amount int64) (string, error)

	/*
		Pay back a payment intent
		@param paymentIntentID - the payment intent id
		@return the payment intent id
	*/
	PayBack(paymentIntentID string) error

	/*
		Find dued payments
		@param userID - the user id
		@return the dued payments
	*/
	FindDuedPayments() ([]types.PaymentInfo, error)
}

type StripePaymentService struct {
	/*
		The Stripe secret key
		@type string
	*/
	StripeSecretKey string
}

func NewStripePaymentService(stripeSecretKey string) *StripePaymentService {
	return &StripePaymentService{
		StripeSecretKey: stripeSecretKey,
	}
}

func (s *StripePaymentService) CreatePaymentIntent(amount int64) (string, error) {
	stripe.Key = s.StripeSecretKey
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		// In the latest version of the API, specifying the `automatic_payment_methods` parameter is optional because Stripe enables its functionality by default.
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

func (s *StripePaymentService) CreateSaveCardIntent() (string, error) {
	stripe.Key = s.StripeSecretKey
	si, _ := setupintent.New(&stripe.SetupIntentParams{
		Usage: stripe.String("off_session"),
	})

	return si.ClientSecret, nil
}

func (s *StripePaymentService) SaveCard(paymentInfo types.PaymentInfo) error {
	// TODO: store to redis?
	return nil
}

func (s *StripePaymentService) Charge(paymentInfo types.PaymentInfo, amount int64) (string, error) {
	stripe.Key = s.StripeSecretKey
	pi, err := paymentintent.New(&stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      stripe.String(paymentInfo.CustomerID),
		PaymentMethod: stripe.String(paymentInfo.PaymentIntentID),
		OffSession:    stripe.Bool(true),
		Confirm:       stripe.Bool(true),
	})

	if err != nil {
		return "", err
	}

	return pi.ClientSecret, nil
}

func (s *StripePaymentService) PayBack(userStripeAccountID string, amount int64) error {
	stripe.Key = s.StripeSecretKey
	payoutParams := &stripe.TransferParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Destination: stripe.String(userStripeAccountID),
	}

	_, err := transfer.New(payoutParams)

	if err != nil {
		return err
	}

	return nil
}
