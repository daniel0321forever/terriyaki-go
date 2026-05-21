package services

import (
	"fmt"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

type providerContractAdapter struct {
	provider entities.PaymentProvider
	prefix   string
}

func (a *providerContractAdapter) CreateCollectionIntent(req CollectionIntentRequestPayload) (CollectionIntentResultPayload, error) {
	switch typed := req.(type) {
	case StripeCollectionIntentRequest:
		ref := fmt.Sprintf("%s_intent_%d", a.prefix, typed.Amount)
		return &StripeCollectionIntentResult{ProviderReference: ref, ClientSecret: ref + "_secret", Status: entities.SettlementStatusPending}, nil
	case SolanaCollectionIntentRequest:
		ref := fmt.Sprintf("%s_intent_%d", a.prefix, typed.Amount)
		return &SolanaCollectionIntentResult{ProviderReference: ref, ClientSecret: ref + "_secret", Status: entities.SettlementStatusPending}, nil
	default:
		return nil, fmt.Errorf("unexpected collection intent request type %T", req)
	}
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

func (a *providerContractAdapter) CreateSettlementIntent(req SettlementIntentRequestPayload) (SettlementIntentResultPayload, error) {
	switch typed := req.(type) {
	case StripeSettlementIntentRequest:
		return &StripeSettlementIntentResult{ProviderReference: fmt.Sprintf("%s_settlement_%s", a.prefix, typed.PaymentMethodID), Status: entities.SettlementStatusCaptured}, nil
	case SolanaSettlementIntentRequest:
		return &SolanaSettlementIntentResult{ProviderReference: fmt.Sprintf("%s_settlement_%s", a.prefix, typed.PaymentMethodID), Status: entities.SettlementStatusCaptured}, nil
	default:
		return nil, fmt.Errorf("unexpected settlement intent request type %T", req)
	}
}

func (a *providerContractAdapter) ResolveSettlement(req SettlementResolutionRequestPayload) (SettlementResolutionResultPayload, error) {
	switch typed := req.(type) {
	case StripeSettlementResolutionRequest:
		return &StripeSettlementResolutionResult{ProviderReference: typed.ProviderReference, Status: typed.Resolution}, nil
	case SolanaSettlementResolutionRequest:
		return &SolanaSettlementResolutionResult{ProviderReference: typed.ProviderReference, Status: entities.SettlementStatus(typed.Resolution), Signature: ""}, nil
	default:
		return nil, fmt.Errorf("unexpected settlement resolution request type %T", req)
	}
}

func (a *providerContractAdapter) QuerySettlementStatus(req QuerySettlementStatusRequestPayload) (SettlementResolutionResultPayload, error) {
	switch typed := req.(type) {
	case StripeQuerySettlementStatusRequest:
		return &StripeSettlementResolutionResult{ProviderReference: typed.ProviderReference, Status: entities.SettlementStatusCaptured}, nil
	case SolanaQuerySettlementStatusRequest:
		return &SolanaSettlementResolutionResult{ProviderReference: typed.ProviderReference, Status: entities.SettlementStatusCaptured}, nil
	default:
		return nil, fmt.Errorf("unexpected query settlement status request type %T", req)
	}
}

func (a *providerContractAdapter) CreateDisbursement(req DisbursementRequestPayload) (DisbursementResultPayload, error) {
	if a.provider == entities.PaymentProviderSolana {
		if _, ok := req.(SolanaDisbursementRequest); !ok {
			return nil, fmt.Errorf("expected SolanaDisbursementRequest, got %T", req)
		}
		return &SolanaDisbursementResult{ProviderReference: a.prefix + "_disbursement", Status: entities.SettlementStatusCaptured}, nil
	}

	if _, ok := req.(StripeDisbursementRequest); !ok {
		return nil, fmt.Errorf("expected StripeDisbursementRequest, got %T", req)
	}
	return &StripeDisbursementResult{ProviderReference: a.prefix + "_disbursement", Status: entities.SettlementStatusCaptured}, nil
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

	if provider == entities.PaymentProviderStripe {
		intentDTO, ctorErr := dto.NewStripeCreateIntentDTO("test-user", 123, "usd")
		if ctorErr != nil {
			t.Fatalf("constructor error: %v", ctorErr)
		}
		intentA, err := svc.CreateStripeCollectionIntent(intentDTO, "intent-key")
		if err != nil {
			t.Fatalf("expected payment intent success, got error: %v", err)
		}
		if intentA.IdempotentReplay {
			t.Fatalf("expected first payment intent call not replayed")
		}

		intentB, err := svc.CreateStripeCollectionIntent(intentDTO, "intent-key")
		if err != nil {
			t.Fatalf("expected payment intent replay success, got error: %v", err)
		}
		if !intentB.IdempotentReplay {
			t.Fatalf("expected second payment intent call replayed")
		}
		if intentA.ClientSecret != intentB.ClientSecret {
			t.Fatalf("expected replayed payment intent response %q, got %q", intentA.ClientSecret, intentB.ClientSecret)
		}
	} else {
		_, err := svc.CreateStripeCollectionIntent(dto.StripeCreateIntentDTO{AmountCents: 123}, "intent-key")
		if err == nil {
			t.Fatalf("expected provider-specific collection intent requirement error for non-stripe providers")
		}
	}

	paymentInfo := entities.PaymentMethodInfo{
		Provider:                provider,
		MethodType:              methodType,
		ProviderCustomerID:      "cust_1",
		ProviderPaymentMethodID: "method_1",
	}

	chargeReqA, ctorErr := dto.NewChargeWithIdempotencyDTO(paymentInfo, 777, "force_charging", "user_1")
	if ctorErr != nil {
		t.Fatalf("constructor error: %v", ctorErr)
	}
	refA, err := svc.ChargeWithIdempotency(chargeReqA, "charge-key")
	if err != nil {
		t.Fatalf("expected settlement success, got error: %v", err)
	}
	if refA.IdempotentReplay {
		t.Fatalf("expected first settlement call not replayed")
	}

	chargeReqB, ctorErr := dto.NewChargeWithIdempotencyDTO(paymentInfo, 777, "force_charging", "user_1")
	if ctorErr != nil {
		t.Fatalf("constructor error: %v", ctorErr)
	}
	refB, err := svc.ChargeWithIdempotency(chargeReqB, "charge-key")
	if err != nil {
		t.Fatalf("expected settlement replay success, got error: %v", err)
	}
	if !refB.IdempotentReplay {
		t.Fatalf("expected second settlement call replayed")
	}
	if refA.ProviderReference != refB.ProviderReference {
		t.Fatalf("expected replayed settlement reference %q, got %q", refA.ProviderReference, refB.ProviderReference)
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
	chargeReq1, ctorErr := dto.NewChargeWithIdempotencyDTO(method, 250, "force_charging", "user_1")
	if ctorErr != nil {
		t.Fatalf("constructor error: %v", ctorErr)
	}
	ref, err := svc.ChargeWithIdempotency(chargeReq1, "solana-method-key")
	if err != nil {
		t.Fatalf("expected settlement success, got error: %v", err)
	}
	if ref.ProviderReference[:6] != "stripe" {
		t.Fatalf("expected stripe adapter reference, got %q", ref.ProviderReference)
	}

	cardMethod := entities.PaymentMethodInfo{
		MethodType:              "card",
		ProviderCustomerID:      "cust_card",
		ProviderPaymentMethodID: "pm_1",
	}
	chargeReq2, ctorErr := dto.NewChargeWithIdempotencyDTO(cardMethod, 250, "force_charging", "user_1")
	if ctorErr != nil {
		t.Fatalf("constructor error: %v", ctorErr)
	}
	ref, err = svc.ChargeWithIdempotency(chargeReq2, "card-method-key")
	if err != nil {
		t.Fatalf("expected stripe settlement success, got error: %v", err)
	}
	if ref.ProviderReference[:6] != "stripe" {
		t.Fatalf("expected stripe adapter reference, got %q", ref.ProviderReference)
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
