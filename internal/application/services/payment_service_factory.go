package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
)

// PaymentProviderDefinition centralizes all configuration for a payment provider:
// metadata (method types) and adapter factory function.
type PaymentProviderDefinition struct {
	MethodTypes  []string
	BuildAdapter func() PaymentGatewayAdapter
}

// providerRegistry is the single source of truth for all payment provider configurations.
// To add a new provider: add an entry here. That's it. Router, factory, and service all use this.
var providerRegistry = map[entities.PaymentProvider]PaymentProviderDefinition{
	entities.PaymentProviderStripe: {
		MethodTypes: []string{"card"},
		BuildAdapter: func() PaymentGatewayAdapter {
			secret := os.Getenv(config.OS_ENV_STRIPE_SECRET_KEY)
			return NewStripePaymentGatewayAdapter(secret)
		},
	},
	entities.PaymentProviderSolana: {
		MethodTypes: []string{"solana_wallet"},
		BuildAdapter: func() PaymentGatewayAdapter {
			return NewSolanaPaymentGatewayAdapter()
		},
	},
}

// PaymentServiceFactory encapsulates PaymentService wiring so route registration
// stays readable while provider selection remains explicit at startup.
type PaymentServiceFactory struct {
	userRepo repositories.UserRepository

	grindRepo         repositories.GrindRepository
	participationRepo repositories.ParticipationRepository

	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository
	idempotencyRepo       repositories.PaymentIdempotencyRepository
	settlementRepo        repositories.PaymentSettlementRepository
}

func NewPaymentServiceFactory(
	userRepo repositories.UserRepository,
	grindRepo repositories.GrindRepository,
	participationRepo repositories.ParticipationRepository,
	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository,
	idempotencyRepo repositories.PaymentIdempotencyRepository,
	settlementRepo repositories.PaymentSettlementRepository,
) *PaymentServiceFactory {
	return &PaymentServiceFactory{
		userRepo:              userRepo,
		grindRepo:             grindRepo,
		participationRepo:     participationRepo,
		paymentMethodInfoRepo: paymentMethodInfoRepo,
		idempotencyRepo:       idempotencyRepo,
		settlementRepo:        settlementRepo,
	}
}

// BuildForProvider creates a payment service bound to the selected provider.
// The provider registry remains the single source of truth for adapter construction.
// This keeps bootstrap logic in one place (factory) and keeps router code clean.
func (f *PaymentServiceFactory) BuildForProvider(
	provider entities.PaymentProvider,
) (*PaymentService, error) {
	if provider == "" {
		return nil, errors.New("payment provider is required")
	}

	// Validate provider exists in registry
	definition, exists := providerRegistry[provider]
	if !exists {
		return nil, fmt.Errorf("payment provider %s not found in provider registry", provider)
	}

	adapter := definition.BuildAdapter()
	if adapter == nil {
		return nil, fmt.Errorf("payment provider %s failed to build adapter", provider)
	}

	svc := newPaymentService(
		f.userRepo,
		f.grindRepo,
		f.participationRepo,
		f.paymentMethodInfoRepo,
		f.idempotencyRepo,
		f.settlementRepo,
		provider,
		adapter,
	)
	if svc == nil {
		return nil, errors.New("failed to create payment service")
	}

	return svc, nil
}
