package services

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/customer"
	"github.com/stripe/stripe-go/v84/paymentintent"
	"github.com/stripe/stripe-go/v84/paymentmethod"
	"github.com/stripe/stripe-go/v84/refund"
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

func (a *StripePaymentGatewayAdapter) CreateCollectionIntent(req CollectionIntentRequest) (*CollectionIntentResult, error) {
	stripe.Key = a.secretKey
	currency := req.Currency
	if currency == "" {
		currency = string(stripe.CurrencyUSD)
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	return &CollectionIntentResult{
		ProviderReference: pi.ID,
		ClientSecret:      pi.ClientSecret,
		Status:            entities.SettlementStatusPending,
	}, nil
}

func (a *StripePaymentGatewayAdapter) CreatePaymentMethodSetupIntent(req PaymentMethodSetupIntentRequest) (*PaymentMethodSetupIntentResult, error) {
	stripe.Key = a.secretKey
	usage := req.Usage
	if usage == "" {
		usage = "off_session"
	}

	si, err := setupintent.New(&stripe.SetupIntentParams{Usage: stripe.String(usage)})
	if err != nil {
		return nil, err
	}

	return &PaymentMethodSetupIntentResult{
		ProviderReference: si.ID,
		ClientSecret:      si.ClientSecret,
	}, nil
}

func (a *StripePaymentGatewayAdapter) EnsurePayerProfile(req PayerProfileRequest) (*PayerProfileResult, error) {
	if req.ExistingPayerReference != "" {
		return &PayerProfileResult{PayerReference: req.ExistingPayerReference}, nil
	}

	stripe.Key = a.secretKey
	params := &stripe.CustomerParams{
		Name:  stripe.String(req.Name),
		Email: stripe.String(req.Email),
	}

	cus, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return &PayerProfileResult{PayerReference: cus.ID}, nil
}

func (a *StripePaymentGatewayAdapter) GetPaymentMethodDetails(paymentMethodID string) (*entities.PaymentMethodInfo, error) {
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

func (a *StripePaymentGatewayAdapter) LinkPaymentMethodToPayer(req PaymentMethodLinkRequest) error {
	stripe.Key = a.secretKey
	attachParams := &stripe.PaymentMethodAttachParams{Customer: stripe.String(req.PayerReference)}
	_, err := paymentmethod.Attach(req.PaymentMethodID, attachParams)
	return err
}

func (a *StripePaymentGatewayAdapter) CreateSettlementIntent(req SettlementIntentRequest) (*SettlementIntentResult, error) {
	stripe.Key = a.secretKey
	currency := req.Currency
	if currency == "" {
		currency = string(stripe.CurrencyUSD)
	}

	pi, err := paymentintent.New(&stripe.PaymentIntentParams{
		Amount:        stripe.Int64(req.Amount),
		Currency:      stripe.String(currency),
		Customer:      stripe.String(req.CustomerID),
		PaymentMethod: stripe.String(req.PaymentMethodID),
		OffSession:    stripe.Bool(true),
		Confirm:       stripe.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	status := entities.SettlementStatusPending
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		status = entities.SettlementStatusCaptured
	}

	return &SettlementIntentResult{
		ProviderReference: pi.ID,
		ClientSecret:      pi.ClientSecret,
		Status:            status,
	}, nil
}

func (a *StripePaymentGatewayAdapter) ResolveSettlement(req SettlementResolutionRequest) (*SettlementResolutionResult, error) {
	stripe.Key = a.secretKey

	if req.Resolution == entities.SettlementStatusRefunded {
		_, err := refund.New(&stripe.RefundParams{PaymentIntent: stripe.String(req.ProviderReference)})
		if err != nil {
			return nil, err
		}
		return &SettlementResolutionResult{ProviderReference: req.ProviderReference, Status: entities.SettlementStatusRefunded}, nil
	}

	return a.QuerySettlementStatus(req.ProviderReference)
}

func (a *StripePaymentGatewayAdapter) QuerySettlementStatus(providerReference string) (*SettlementResolutionResult, error) {
	stripe.Key = a.secretKey
	pi, err := paymentintent.Get(providerReference, nil)
	if err != nil {
		return nil, err
	}

	status := entities.SettlementStatusPending
	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded:
		status = entities.SettlementStatusCaptured
	case stripe.PaymentIntentStatusRequiresCapture:
		status = entities.SettlementStatusAuthorized
	case stripe.PaymentIntentStatusCanceled:
		status = entities.SettlementStatusFailed
	}

	return &SettlementResolutionResult{ProviderReference: providerReference, Status: status}, nil
}

func (a *StripePaymentGatewayAdapter) CreateDisbursement(req DisbursementRequest) (*DisbursementResult, error) {
	stripe.Key = a.secretKey
	currency := req.Currency
	if currency == "" {
		currency = string(stripe.CurrencyUSD)
	}

	_, err := transfer.New(&stripe.TransferParams{
		Amount:      stripe.Int64(req.Amount),
		Currency:    stripe.String(currency),
		Destination: stripe.String(req.DestinationReference),
	})
	if err != nil {
		return nil, err
	}

	return &DisbursementResult{ProviderReference: req.DestinationReference, Status: entities.SettlementStatusCaptured}, nil
}
