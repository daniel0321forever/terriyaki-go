package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"github.com/redis/go-redis/v9"
)

// Payment service interface
type IPaymentService interface {
	// Intent lifecycle
	CreatePaymentIntentWithIdempotency(amount int64, idempotencyKey string) (string, bool, error)

	// Payment method lifecycle
	AddPaymentMethod(request dto.AddPaymentMethodDTO) (*entities.PaymentMethodInfo, error)

	// Charging lifecycle
	ChargeWithIdempotency(paymentInfo entities.PaymentMethodInfo, amount int64, operation string, idempotencyKey string, userID string) (string, bool, error)

	PayBack(destinationAccountID string, amount int64) error

	// Settlement/reconciliation
	FindDuedPayments() ([]dto.PendingPaymentDTO, error)
	ReconcileSettlements(limit int) ([]entities.PaymentSettlement, error)

	// Queries and idempotency helpers
	GetAvailablePaymentMethods(request dto.GetAvailablePaymentMethodsDTO) (*dto.AvailablePaymentMethodsDTO, error)
	ClaimIdempotency(operation string, idempotencyKey string) (bool, error)
}

type PaymentService struct {
	provider entities.PaymentProvider
	adapter  PaymentGatewayAdapter
	cardAdapter   CardMethodAdapter
	walletAdapter WalletMethodAdapter
	// RedisClient is currently kept for future queue/retry usage.
	RedisClient *redis.Client

	userRepo          repositories.UserRepository
	grindRepo         repositories.GrindRepository
	participationRepo repositories.ParticipationRepository

	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository
	idempotencyRepo       repositories.PaymentIdempotencyRepository
	settlementRepo        repositories.PaymentSettlementRepository
}

func newPaymentService(
	userRepo repositories.UserRepository,
	grindRepo repositories.GrindRepository,
	participationRepo repositories.ParticipationRepository,
	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository,
	idempotencyRepo repositories.PaymentIdempotencyRepository,
	settlementRepo repositories.PaymentSettlementRepository,
	provider entities.PaymentProvider,
	adapter PaymentGatewayAdapter,
) *PaymentService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})

	if provider == "" || adapter == nil {
		return nil
	}

	svc := &PaymentService{
		provider:              provider,
		adapter:               adapter,
		RedisClient:           rdb,
		userRepo:              userRepo,
		grindRepo:             grindRepo,
		participationRepo:     participationRepo,
		paymentMethodInfoRepo: paymentMethodInfoRepo,
		idempotencyRepo:       idempotencyRepo,
		settlementRepo:        settlementRepo,
	}

	if cardAdapter, ok := any(adapter).(CardMethodAdapter); ok {
		svc.cardAdapter = cardAdapter
	}
	if walletAdapter, ok := any(adapter).(WalletMethodAdapter); ok {
		svc.walletAdapter = walletAdapter
	}

	return svc
}

func (s *PaymentService) CreatePaymentIntentWithIdempotency(amount int64, idempotencyKey string) (string, bool, error) {

	claimed, err := s.ClaimIdempotency("payment_intent", idempotencyKey)
	if err != nil {
		return "", false, err
	}
	if !claimed {
		if s.idempotencyRepo == nil {
			return "", true, nil
		}
		response, getErr := s.idempotencyRepo.GetResponse("payment_intent", idempotencyKey)
		if getErr != nil {
			return "", true, nil
		}
		return response, true, nil
	}

	intent, err := s.adapter.CreateCollectionIntent(CollectionIntentRequest{Amount: amount, Currency: "usd"})
	if err != nil {
		return "", false, err
	}
	clientSecret := intent.ClientSecret

	if s.idempotencyRepo != nil {
		_ = s.idempotencyRepo.SetResponse("payment_intent", idempotencyKey, clientSecret)
	}

	return clientSecret, false, nil
}

func (s *PaymentService) AddPaymentMethod(
	request dto.AddPaymentMethodDTO,
) (*entities.PaymentMethodInfo, error) {
	if request.MethodType == "card" {
		return s.addCardMethod(request)
	}
	if request.MethodType == "solana_wallet" {
		return s.addSolanaWalletMethod(request)
	}
	return nil, fmt.Errorf("unsupported payment method type: %s", request.MethodType)
}

func (s *PaymentService) addCardMethod(request dto.AddPaymentMethodDTO) (*entities.PaymentMethodInfo, error) {
	if s.provider != entities.PaymentProviderStripe {
		return nil, fmt.Errorf("card payment method requires Stripe provider")
	}
	if s.cardAdapter == nil {
		return nil, fmt.Errorf("card onboarding is not available for provider %s", s.provider)
	}
	if request.CardPaymentMethodID == "" {
		return nil, fmt.Errorf("card payment method ID is required")
	}

	// create payer profile if not exists
	var payerReference string

	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return nil, err
	}

	profile, err := s.cardAdapter.EnsurePayerProfile(PayerProfileRequest{
		Name:                   user.Username,
		Email:                  user.Email,
		ExistingPayerReference: user.StripeCustomerID,
	})
	if err != nil {
		return nil, err
	}

	payerReference = profile.PayerReference
	if user.StripeCustomerID == "" {
		fmt.Print("created payer profile id " + payerReference + " for user " + user.Username)
		user.StripeCustomerID = payerReference
		user.DefaultPaymentMethodID = request.CardPaymentMethodID
		err = s.userRepo.Update(user)
		if err != nil {
			return nil, err
		}
	}

	// link the payment method to payer profile
	err = s.cardAdapter.LinkPaymentMethodToPayer(PaymentMethodLinkRequest{
		PaymentMethodID: request.CardPaymentMethodID,
		PayerReference:  payerReference,
	})
	if err != nil {
		return nil, err
	}

	paymentMethodInfo, err := s.cardAdapter.GetPaymentMethodDetails(request.CardPaymentMethodID)
	if err != nil {
		return nil, err
	}
	paymentMethodInfo.UserID = user.ID
	paymentMethodInfo.ProviderCustomerID = payerReference
	paymentMethodInfo.Provider = s.provider

	paymentMethodInfo, err = s.paymentMethodInfoRepo.Create(paymentMethodInfo)
	if err != nil {
		return nil, err
	}
	return paymentMethodInfo, nil
}

func (s *PaymentService) addSolanaWalletMethod(request dto.AddPaymentMethodDTO) (*entities.PaymentMethodInfo, error) {
	if s.provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("solana wallet requires Solana provider")
	}
	if request.WalletAddress == "" {
		return nil, fmt.Errorf("solana wallet address is required")
	}
	if request.Network == "" {
		return nil, fmt.Errorf("solana network is required")
	}

	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return nil, err
	}

	walletRequest := WalletMethodRequest{
		UserID:        user.ID,
		WalletAddress: request.WalletAddress,
		Network:       request.Network,
		ProgramID:     request.ProgramID,
	}

	if s.walletAdapter == nil {
		return nil, fmt.Errorf("wallet onboarding is not available for provider %s", s.provider)
	}
	if err := s.walletAdapter.ValidateWalletOwnership(walletRequest); err != nil {
		return nil, err
	}

	paymentMethodInfo, err := s.walletAdapter.NormalizeWalletMethod(walletRequest)
	if err != nil {
		return nil, err
	}
	paymentMethodInfo.UserID = user.ID
	paymentMethodInfo.Provider = s.provider

	paymentMethodInfo, err = s.paymentMethodInfoRepo.Create(paymentMethodInfo)
	if err != nil {
		return nil, err
	}
	return paymentMethodInfo, nil
}

func (s *PaymentService) ChargeWithIdempotency(paymentInfo entities.PaymentMethodInfo, amount int64, operation string, idempotencyKey string, userID string) (string, bool, error) {

	claimed, err := s.ClaimIdempotency(operation, idempotencyKey)
	if err != nil {
		return "", false, err
	}
	if !claimed {
		if s.settlementRepo == nil {
			return "", true, nil
		}
		settlement, findErr := s.settlementRepo.FindByOperationAndKey(operation, idempotencyKey)
		if findErr != nil {
			return "", true, nil
		}
		return settlement.Reference.ProviderReference, true, nil
	}

	provider := s.provider

	var settlement *entities.PaymentSettlement
	if s.settlementRepo != nil {
		settlement = entities.NewPaymentSettlement(userID, operation, idempotencyKey, provider, paymentInfo.ProviderPaymentMethodID, amount)
		settlement.Reference.Network = paymentInfo.Network
		settlement, err = s.settlementRepo.Create(settlement)
		if err != nil {
			return "", false, err
		}
	}

	intent, chargeErr := s.adapter.CreateSettlementIntent(SettlementIntentRequest{
		CustomerID:      paymentInfo.ProviderCustomerID,
		PaymentMethodID: paymentInfo.ProviderPaymentMethodID,
		Amount:          amount,
		Currency:        "usd",
	})
	if chargeErr != nil {
		if settlement != nil {
			settlement.Status = entities.SettlementStatusFailed
			settlement.LastError = chargeErr.Error()
			_, _ = s.settlementRepo.Update(settlement)
		}
		return "", false, chargeErr
	}

	reference := intent.ProviderReference

	if settlement != nil {
		settlement.Status = intent.Status
		if settlement.Status == "" {
			settlement.Status = entities.SettlementStatusCaptured
		}
		settlement.Reference.ProviderReference = reference
		_, _ = s.settlementRepo.Update(settlement)
	}

	if s.idempotencyRepo != nil {
		_ = s.idempotencyRepo.SetResponse(operation, idempotencyKey, reference)
	}

	return reference, false, nil
}

func (s *PaymentService) PayBack(destinationAccountID string, amount int64) error {
	_, err := s.adapter.CreateDisbursement(DisbursementRequest{DestinationReference: destinationAccountID, Amount: amount, Currency: "usd"})
	return err
}

func (s *PaymentService) GetAvailablePaymentMethods(request dto.GetAvailablePaymentMethodsDTO) (*dto.AvailablePaymentMethodsDTO, error) {
	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// get payment infos
	paymentInfos, err := s.paymentMethodInfoRepo.FindByUserID(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default payment method: %w", err)
	}

	// get default one
	var defaultPaymentInfo entities.PaymentMethodInfo
	for _, paymentInfo := range paymentInfos {
		if paymentInfo.ProviderPaymentMethodID == user.DefaultPaymentMethodID {
			defaultPaymentInfo = paymentInfo
			break
		}
	}

	return &dto.AvailablePaymentMethodsDTO{
		PaymentInfos:       paymentInfos,
		DefaultPaymentInfo: defaultPaymentInfo,
	}, nil
}

func (s *PaymentService) FindDuedPayments() ([]dto.PendingPaymentDTO, error) {
	var grinds []*entities.Grind

	// Find all grinds where StartDate + Duration (in days) is the current date (today, UTC)
	grinds, err := s.grindRepo.FindDuedGrinds()
	if err != nil {
		return nil, err
	}

	// get the punishment for each grind
	var pendingPayments []dto.PendingPaymentDTO

	for _, g := range grinds {
		for _, p := range g.Participants {
			// get the stripe payment info for the user
			var paymentMethodInfos []entities.PaymentMethodInfo
			paymentMethodInfos, err := s.paymentMethodInfoRepo.FindByUserID(p.ID)
			if err != nil {
				return nil, err
			}

			// get the total penalty for the user in the grind
			participateRecord, err := s.participationRepo.FindByUserAndGrind(p.ID, g.ID)
			if err != nil {
				return nil, err
			}

			pendingPayments = append(pendingPayments, dto.PendingPaymentDTO{
				PaymentMethodInfo: paymentMethodInfos[0],
				PaymentAmount:     int64(participateRecord.TotalPenalty),
			})
		}
	}

	return pendingPayments, nil
}

func (s *PaymentService) ReconcileSettlements(limit int) ([]entities.PaymentSettlement, error) {
	if s.settlementRepo == nil {
		return []entities.PaymentSettlement{}, nil
	}

	settlements, err := s.settlementRepo.FindByStatuses([]entities.SettlementStatus{entities.SettlementStatusPending, entities.SettlementStatusFailed}, limit)
	if err != nil {
		return nil, err
	}

	updated := make([]entities.PaymentSettlement, 0, len(settlements))
	nowUnix := time.Now().Unix()

	for i := range settlements {
		settlement := settlements[i]
		shouldUpdate := false

		switch settlement.Status {
		case entities.SettlementStatusPending:
			if nowUnix-settlement.CreatedAt.Unix() > 600 {
				settlement.Status = entities.SettlementStatusFailed
				settlement.LastError = "settlement timed out during reconciliation"
				shouldUpdate = true
			}
		case entities.SettlementStatusFailed:
			if settlement.RetryCount < 3 {
				settlement.RetryCount++
				settlement.Status = entities.SettlementStatusPending
				settlement.LastError = "retry scheduled by reconciliation"
				shouldUpdate = true
			}
		}

		if shouldUpdate {
			updatedSettlement, updateErr := s.settlementRepo.Update(&settlement)
			if updateErr != nil {
				return nil, updateErr
			}
			updated = append(updated, *updatedSettlement)
		}
	}

	return updated, nil
}

func (s *PaymentService) ClaimIdempotency(operation string, idempotencyKey string) (bool, error) {
	if idempotencyKey == "" {
		return false, errors.New("idempotency key is required")
	}
	if s.idempotencyRepo == nil {
		return true, nil
	}
	return s.idempotencyRepo.Claim(operation, idempotencyKey)
}
