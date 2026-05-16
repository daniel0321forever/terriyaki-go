package services

import (
	"fmt"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

type providerContractAdapter struct {
	provider entities.PaymentProvider
	prefix   string
}

func (a *providerContractAdapter) CreateCollectionIntent(req CollectionIntentRequest) (*CollectionIntentResult, error) {
	ref := fmt.Sprintf("%s_intent_%d", a.prefix, req.Amount)
	return &CollectionIntentResult{
		ProviderReference: ref,
		ClientSecret:      ref + "_secret",
		Status:            entities.SettlementStatusPending,
	}, nil
}

func (a *providerContractAdapter) CreatePaymentMethodSetupIntent(req PaymentMethodSetupIntentRequest) (*PaymentMethodSetupIntentResult, error) {
	ref := a.prefix + "_setup"
	return &PaymentMethodSetupIntentResult{ProviderReference: ref, ClientSecret: ref + "_secret"}, nil
}

func (a *providerContractAdapter) EnsurePayerProfile(req PayerProfileRequest) (*PayerProfileResult, error) {
	if req.ExistingPayerReference != "" {
		return &PayerProfileResult{PayerReference: req.ExistingPayerReference}, nil
	}
	return &PayerProfileResult{PayerReference: a.prefix + "_payer"}, nil
}

func (a *providerContractAdapter) GetPaymentMethodDetails(paymentMethodID string) (*entities.PaymentMethodInfo, error) {
	info := entities.NewPaymentMethodInfo(a.provider, "card", "", a.prefix+"_payer", paymentMethodID, "visa", "4242", 1, 2030)
	if a.provider == entities.PaymentProviderSolana {
		info.MethodType = "solana_wallet"
		info.Network = "solana"
		info.WalletAddress = "wallet_123"
	}
	return info, nil
}

func (a *providerContractAdapter) LinkPaymentMethodToPayer(req PaymentMethodLinkRequest) error {
	return nil
}

func (a *providerContractAdapter) CreateSettlementIntent(req SettlementIntentRequest) (*SettlementIntentResult, error) {
	return &SettlementIntentResult{
		ProviderReference: fmt.Sprintf("%s_settlement_%s", a.prefix, req.PaymentMethodID),
		Status:            entities.SettlementStatusCaptured,
	}, nil
}

func (a *providerContractAdapter) ResolveSettlement(req SettlementResolutionRequest) (*SettlementResolutionResult, error) {
	return &SettlementResolutionResult{ProviderReference: req.ProviderReference, Status: req.Resolution}, nil
}

func (a *providerContractAdapter) QuerySettlementStatus(providerReference string) (*SettlementResolutionResult, error) {
	return &SettlementResolutionResult{ProviderReference: providerReference, Status: entities.SettlementStatusCaptured}, nil
}

func (a *providerContractAdapter) CreateDisbursement(req DisbursementRequest) (*DisbursementResult, error) {
	return &DisbursementResult{ProviderReference: a.prefix + "_disbursement", Status: entities.SettlementStatusCaptured}, nil
}

func runSwappableAdapterSuite(t *testing.T, provider entities.PaymentProvider, methodType string, adapter PaymentGatewayAdapter) {
	t.Helper()

	idempotencyRepo := newInMemoryIdempotencyRepo()
	settlementRepo := newInMemorySettlementRepo()
	svc := newPaymentService(
		nil,
		nil,
		nil,
		nil,
		idempotencyRepo,
		settlementRepo,
		provider,
		adapter,
	)
	if svc == nil {
		t.Fatalf("expected payment service, got nil")
	}

	intentA, replayed, err := svc.CreatePaymentIntentWithIdempotency(123, "intent-key")
	if err != nil {
		t.Fatalf("expected payment intent success, got error: %v", err)
	}
	if replayed {
		t.Fatalf("expected first payment intent call not replayed")
	}

	intentB, replayed, err := svc.CreatePaymentIntentWithIdempotency(123, "intent-key")
	if err != nil {
		t.Fatalf("expected payment intent replay success, got error: %v", err)
	}
	if !replayed {
		t.Fatalf("expected second payment intent call replayed")
	}
	if intentA != intentB {
		t.Fatalf("expected replayed payment intent response %q, got %q", intentA, intentB)
	}

	paymentInfo := entities.PaymentMethodInfo{
		Provider:                provider,
		MethodType:              methodType,
		ProviderCustomerID:      "cust_1",
		ProviderPaymentMethodID: "method_1",
	}

	refA, replayed, err := svc.ChargeWithIdempotency(paymentInfo, 777, "force_charging", "charge-key", "user_1")
	if err != nil {
		t.Fatalf("expected settlement success, got error: %v", err)
	}
	if replayed {
		t.Fatalf("expected first settlement call not replayed")
	}

	refB, replayed, err := svc.ChargeWithIdempotency(paymentInfo, 777, "force_charging", "charge-key", "user_1")
	if err != nil {
		t.Fatalf("expected settlement replay success, got error: %v", err)
	}
	if !replayed {
		t.Fatalf("expected second settlement call replayed")
	}
	if refA != refB {
		t.Fatalf("expected replayed settlement reference %q, got %q", refA, refB)
	}

	stored, err := settlementRepo.FindByOperationAndKey("force_charging", "charge-key")
	if err != nil {
		t.Fatalf("expected stored settlement, got error: %v", err)
	}
	if stored.Provider != provider {
		t.Fatalf("expected provider %q, got %q", provider, stored.Provider)
	}
	if stored.Status != entities.SettlementStatusCaptured {
		t.Fatalf("expected captured settlement status, got %q", stored.Status)
	}
}

func TestAdapterSelectionByMethodType(t *testing.T) {
	t.Parallel()

	stripeAdapter := &providerContractAdapter{provider: entities.PaymentProviderStripe, prefix: "stripe"}

	idempotencyRepo := newInMemoryIdempotencyRepo()
	settlementRepo := newInMemorySettlementRepo()
	svc := newPaymentService(
		nil,
		nil,
		nil,
		nil,
		idempotencyRepo,
		settlementRepo,
		entities.PaymentProviderStripe,
		stripeAdapter,
	)
	if svc == nil {
		t.Fatalf("expected payment service, got nil")
	}

	method := entities.PaymentMethodInfo{
		MethodType:              "solana_wallet",
		ProviderCustomerID:      "cust_sol",
		ProviderPaymentMethodID: "wallet_1",
		Network:                 "solana",
	}
	ref, _, err := svc.ChargeWithIdempotency(method, 250, "force_charging", "solana-method-key", "user_1")
	if err != nil {
		t.Fatalf("expected settlement success, got error: %v", err)
	}
	if ref[:6] != "stripe" {
		t.Fatalf("expected stripe adapter reference, got %q", ref)
	}

	cardMethod := entities.PaymentMethodInfo{
		MethodType:              "card",
		ProviderCustomerID:      "cust_card",
		ProviderPaymentMethodID: "pm_1",
	}
	ref, _, err = svc.ChargeWithIdempotency(cardMethod, 250, "force_charging", "card-method-key", "user_1")
	if err != nil {
		t.Fatalf("expected stripe settlement success, got error: %v", err)
	}
	if ref[:6] != "stripe" {
		t.Fatalf("expected stripe adapter reference, got %q", ref)
	}
}

func TestSwappableAdapterContractSuite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		provider   entities.PaymentProvider
		methodType string
		adapter    PaymentGatewayAdapter
	}{
		{
			name:       "stripe",
			provider:   entities.PaymentProviderStripe,
			methodType: "card",
			adapter:    &providerContractAdapter{provider: entities.PaymentProviderStripe, prefix: "stripe"},
		},
		{
			name:       "solana",
			provider:   entities.PaymentProviderSolana,
			methodType: "solana_wallet",
			adapter:    &providerContractAdapter{provider: entities.PaymentProviderSolana, prefix: "solana"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runSwappableAdapterSuite(t, tc.provider, tc.methodType, tc.adapter)
		})
	}
}
