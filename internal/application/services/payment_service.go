package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	controlsolana "github.com/daniel0321forever/terriyaki-go/control/solana"
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	solana "github.com/gagliardetto/solana-go"
	"github.com/redis/go-redis/v9"
)

// PaymentServiceCore captures shared payment orchestration that is not tied to one provider.
type PaymentServiceCore interface {
	AddPaymentMethod(request dto.AddPaymentMethodDTO) (*dto.AddPaymentMethodResultDTO, error)
	ChargeWithIdempotency(request dto.ChargeWithIdempotencyDTO, idempotencyKey string) (*dto.ChargeWithIdempotencyResultDTO, error)
	PayBack(request dto.PayBackDTO) (*dto.PayBackResultDTO, error)
	FindDuedPayments() (*dto.PendingPaymentsResultDTO, error)
	ReconcileSettlements(request dto.ReconcileSettlementsDTO) (*dto.ReconcileSettlementsResultDTO, error)
	GetAvailablePaymentMethods(request dto.GetAvailablePaymentMethodsDTO) (*dto.AvailablePaymentMethodsDTO, error)
	ClaimIdempotency(operation string, idempotencyKey string) (*dto.ClaimIdempotencyResultDTO, error)
}

// StripePaymentService exposes Stripe-specific entrypoints plus shared payment orchestration.
type StripePaymentService interface {
	PaymentServiceCore
	CreateStripeCollectionIntent(request dto.StripeCreateIntentDTO, idempotencyKey string) (*dto.StripeCreateCollectionIntentResultDTO, error)
}

// SolanaPaymentService exposes Solana-specific entrypoints plus shared payment orchestration.
type SolanaPaymentService interface {
	PaymentServiceCore
	CreateSolanaCollectionIntent(request dto.SolanaCreateIntentDTO, idempotencyKey string) (*dto.SolanaCreateCollectionIntentResultDTO, error)
	SubmitSolanaSignedTransaction(request dto.SolanaSubmitSignedTransactionDTO, idempotencyKey string) (*dto.SolanaSubmitSignedTransactionResultDTO, error)
}

type solanaConfigProvider interface {
	RPCEndpoint() string
	ProgramID() [32]byte
	OraclePubkey() [32]byte
	OraclePrivateKey() [64]byte
}

type PaymentService struct {
	provider      entities.PaymentProvider
	adapter       PaymentGatewayAdapter
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

// ------------------------------------------------------------
// Shared PaymentService methods (provider-agnostic orchestration)
// These methods are used regardless of whether the underlying
// provider is Stripe or Solana.
// ------------------------------------------------------------

func (s *PaymentService) AddPaymentMethod(
	request dto.AddPaymentMethodDTO,
) (*dto.AddPaymentMethodResultDTO, error) {
	if request.MethodType == "card" {
		return s.addStripeCardMethod(request)
	}
	if request.MethodType == "solana_wallet" {
		return s.addSolanaWalletMethod(request)
	}
	return nil, fmt.Errorf("unsupported payment method type: %s", request.MethodType)
}

func (s *PaymentService) ChargeWithIdempotency(request dto.ChargeWithIdempotencyDTO, idempotencyKey string) (*dto.ChargeWithIdempotencyResultDTO, error) {

	paymentInfo := request.PaymentMethodInfo
	amount := request.AmountCents
	operation := request.Operation
	userID := request.UserID

	res, err := s.ClaimIdempotency(operation, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !res.Claimed {
		if s.settlementRepo == nil {
			return &dto.ChargeWithIdempotencyResultDTO{ProviderReference: "", IdempotentReplay: true}, nil
		}
		settlement, findErr := s.settlementRepo.FindByOperationAndKey(operation, idempotencyKey)
		if findErr != nil {
			return &dto.ChargeWithIdempotencyResultDTO{ProviderReference: "", IdempotentReplay: true}, nil
		}
		return &dto.ChargeWithIdempotencyResultDTO{ProviderReference: settlement.Reference.ProviderReference, IdempotentReplay: true}, nil
	}

	provider := s.provider

	var settlement *entities.PaymentSettlement
	if s.settlementRepo != nil {
		settlement = entities.NewPaymentSettlement(userID, operation, idempotencyKey, provider, paymentInfo.ProviderPaymentMethodID, amount)
		settlement.Reference.Network = paymentInfo.Network
		settlement, err = s.settlementRepo.Create(settlement)
		if err != nil {
			return nil, err
		}
	}

	settlementDTO, dtoErr := dto.NewSettlementIntentRequestDTO(paymentInfo, amount)
	if dtoErr != nil {
		return nil, dtoErr
	}
	settlementReq, reqErr := s.buildSettlementIntentRequest(settlementDTO)
	if reqErr != nil {
		return nil, reqErr
	}

	intentPayload, chargeErr := s.adapter.CreateSettlementIntent(settlementReq)
	if chargeErr != nil {
		if settlement != nil {
			settlement.Status = entities.SettlementStatusFailed
			settlement.LastError = chargeErr.Error()
			_, _ = s.settlementRepo.Update(settlement)
		}
		return nil, chargeErr
	}

	reference, status, extractErr := extractReferenceAndStatusFromSettlementIntent(intentPayload)
	if extractErr != nil {
		if settlement != nil {
			settlement.Status = entities.SettlementStatusFailed
			settlement.LastError = extractErr.Error()
			_, _ = s.settlementRepo.Update(settlement)
		}
		return nil, extractErr
	}

	if settlement != nil {
		settlement.Status = status
		if settlement.Status == "" {
			settlement.Status = entities.SettlementStatusCaptured
		}
		settlement.Reference.ProviderReference = reference
		_, _ = s.settlementRepo.Update(settlement)
	}

	s.setStoredResponse(operation, idempotencyKey, reference)

	return &dto.ChargeWithIdempotencyResultDTO{ProviderReference: reference, IdempotentReplay: false}, nil
}

func (s *PaymentService) PayBack(request dto.PayBackDTO) (*dto.PayBackResultDTO, error) {
	disbursementReq, reqErr := s.buildDisbursementRequest(request)
	if reqErr != nil {
		return nil, reqErr
	}

	result, err := s.adapter.CreateDisbursement(disbursementReq)
	if err != nil {
		return nil, err
	}

	providerRef, status, extractErr := extractReferenceAndStatusFromDisbursement(result)
	if extractErr != nil {
		return nil, extractErr
	}

	return &dto.PayBackResultDTO{
		ProviderReference: providerRef,
		Status:            string(status),
	}, nil
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

func (s *PaymentService) FindDuedPayments() (*dto.PendingPaymentsResultDTO, error) {
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

	return &dto.PendingPaymentsResultDTO{PendingPayments: pendingPayments}, nil
}

func (s *PaymentService) ReconcileSettlements(request dto.ReconcileSettlementsDTO) (*dto.ReconcileSettlementsResultDTO, error) {
	limit := request.Limit

	if s.settlementRepo == nil {
		return &dto.ReconcileSettlementsResultDTO{UpdatedSettlements: []dto.PaymentSettlementDTO{}}, nil
	}

	settlements, err := s.settlementRepo.FindByStatuses([]entities.SettlementStatus{entities.SettlementStatusPending, entities.SettlementStatusFailed}, limit)
	if err != nil {
		return nil, err
	}

	updated := make([]dto.PaymentSettlementDTO, 0, len(settlements))
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
			updated = append(updated, buildPaymentSettlementDTO(updatedSettlement))
		}
	}

	return &dto.ReconcileSettlementsResultDTO{UpdatedSettlements: updated}, nil
}

// ResolvePledgeAsOracle allows the backend to sign and submit a resolution transaction
// using the oracle private key. This implements the oracle-signed path in the hybrid model.
func (s *PaymentService) ResolvePledgeAsOracle(request dto.SolanaResolvePledgeDTO, idempotencyKey string) (*dto.SolanaResolvePledgeResultDTO, error) {
	if s.provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("oracle resolution requires Solana provider")
	}

	claimed, stored, err := s.claimOrGetStoredResponse("solana_resolve_pledge", idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !claimed {
		if stored != "" {
			var resp dto.SolanaResolvePledgeResultDTO
			if unmarshalErr := json.Unmarshal([]byte(stored), &resp); unmarshalErr == nil {
				return &resp, nil
			}
		}
	}

	resolutionReq, reqErr := s.buildSolanaResolvePledgeRequest(request)
	if reqErr != nil {
		return nil, reqErr
	}

	// Retrieve settlement by operation + idempotency key if exists
	var settlement *entities.PaymentSettlement
	if s.settlementRepo != nil {
		settlement, _ = s.settlementRepo.FindByOperationAndKey(request.Operation, idempotencyKey)
	}

	// Delegate signing and submission to the adapter
	resultPayload, adapterErr := s.adapter.ResolveSettlement(resolutionReq)
	if adapterErr != nil {
		if settlement != nil {
			settlement.Status = entities.SettlementStatusFailed
			settlement.LastError = adapterErr.Error()
			_, _ = s.settlementRepo.Update(settlement)
		}
		return nil, adapterErr
	}

	result, ok := resultPayload.(*SolanaSettlementResolutionResult)
	if !ok {
		return nil, fmt.Errorf("unexpected solana settlement resolution result type %T", resultPayload)
	}

	// Update settlement with the result
	if settlement != nil {
		settlement.Status = result.Status
		settlement.Reference.ProviderReference = request.Operation
		settlement.Reference.TxHash = result.Signature
		settlement.Reference.SettlementProof = result.SettlementProof
		settlement.Reference.FinalizedAtUnix = time.Now().Unix()
		settlement.LastError = ""
		_, _ = s.settlementRepo.Update(settlement)
	}

	response := &dto.SolanaResolvePledgeResultDTO{
		ProviderReference: request.Operation,
		Signature:         result.Signature,
		Status:            string(result.Status),
		Resolution:        request.Resolution,
		SettlementProof:   result.SettlementProof,
	}

	if encoded, marshalErr := json.Marshal(response); marshalErr == nil {
		s.setStoredResponse("solana_resolve_pledge", idempotencyKey, string(encoded))
	}

	return response, nil
}

func (s *PaymentService) ClaimIdempotency(operation string, idempotencyKey string) (*dto.ClaimIdempotencyResultDTO, error) {
	if idempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}
	if s.idempotencyRepo == nil {
		return &dto.ClaimIdempotencyResultDTO{Claimed: true}, nil
	}
	claimed, err := s.idempotencyRepo.Claim(operation, idempotencyKey)
	if err != nil {
		return nil, err
	}
	return &dto.ClaimIdempotencyResultDTO{Claimed: claimed}, nil
}

// claimOrGetStoredResponse centralizes idempotency claim + stored-response lookup.
// Returns (claimed, storedResponseString, error).
//   - claimed=false indicates another process has claimed the op; callers may
//     attempt to use the stored response (if non-empty) or perform a provider-specific replay.
//   - claimed=true means the caller should proceed to perform the operation and
//     then call setStoredResponse to persist the result for future replays.
func (s *PaymentService) claimOrGetStoredResponse(operation string, idempotencyKey string) (bool, string, error) {
	if idempotencyKey == "" {
		return true, "", nil
	}
	if s.idempotencyRepo == nil {
		return true, "", nil
	}
	claimed, err := s.idempotencyRepo.Claim(operation, idempotencyKey)
	if err != nil {
		return false, "", err
	}
	if !claimed {
		stored, getErr := s.idempotencyRepo.GetResponse(operation, idempotencyKey)
		if getErr == nil && stored != "" {
			return false, stored, nil
		}
		return false, "", nil
	}
	return true, "", nil
}

func (s *PaymentService) setStoredResponse(operation string, idempotencyKey string, response string) {
	if s.idempotencyRepo != nil && idempotencyKey != "" {
		_ = s.idempotencyRepo.SetResponse(operation, idempotencyKey, response)
	}
}

func (s *PaymentService) buildStripeCollectionIntentRequest(request dto.StripeCreateIntentDTO) (CollectionIntentRequestPayload, error) {
	if s.provider != entities.PaymentProviderStripe {
		return nil, fmt.Errorf("stripe collection intent requires Stripe provider")
	}
	if request.AmountCents <= 0 {
		return nil, fmt.Errorf("amount_cents must be positive for stripe collection intent")
	}
	currency := request.Currency
	if currency == "" {
		currency = "usd"
	}
	return StripeCollectionIntentRequest{Amount: request.AmountCents, Currency: currency}, nil
}

func (s *PaymentService) buildSettlementIntentRequest(request dto.SettlementIntentRequestDTO) (SettlementIntentRequestPayload, error) {
	paymentInfo := request.PaymentMethodInfo
	amount := request.AmountCents
	switch s.provider {
	case entities.PaymentProviderStripe:
		return StripeSettlementIntentRequest{
			CustomerID:      paymentInfo.ProviderCustomerID,
			PaymentMethodID: paymentInfo.ProviderPaymentMethodID,
			Amount:          amount,
			Currency:        "usd",
		}, nil
	case entities.PaymentProviderSolana:
		return s.buildSolanaSettlementIntentRequest(request)
	default:
		return nil, fmt.Errorf("unsupported provider for settlement intent: %s", s.provider)
	}
}

func (s *PaymentService) buildSolanaSettlementIntentRequest(request dto.SettlementIntentRequestDTO) (SettlementIntentRequestPayload, error) {
	paymentInfo := request.PaymentMethodInfo
	amount := request.AmountCents
	paymentMethodID := strings.TrimSpace(paymentInfo.ProviderPaymentMethodID)
	if paymentMethodID == "" {
		paymentMethodID = strings.TrimSpace(paymentInfo.WalletAddress)
	}
	if paymentMethodID == "" {
		return nil, fmt.Errorf("solana payment method reference is required")
	}
	if paymentInfo.Provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("solana settlement intent requires a solana payment method")
	}

	return SolanaSettlementIntentRequest{
		CustomerID:      paymentInfo.ProviderCustomerID,
		PaymentMethodID: paymentMethodID,
		Amount:          amount,
		Currency:        "usd",
	}, nil
}

func (s *PaymentService) buildDisbursementRequest(request dto.PayBackDTO) (DisbursementRequestPayload, error) {
	switch s.provider {
	case entities.PaymentProviderStripe:
		return StripeDisbursementRequest{DestinationReference: request.DestinationAccountID, Amount: request.AmountCents, Currency: "usd"}, nil
	case entities.PaymentProviderSolana:
		if strings.TrimSpace(request.DestinationAccountID) == "" {
			return nil, fmt.Errorf("solana disbursement destination is required")
		}
		if request.AmountCents <= 0 {
			return nil, fmt.Errorf("solana disbursement amount must be greater than zero")
		}
		return SolanaDisbursementRequest{DestinationReference: request.DestinationAccountID, Amount: request.AmountCents, Currency: "usd"}, nil
	default:
		return nil, fmt.Errorf("unsupported provider for disbursement: %s", s.provider)
	}
}

func (s *PaymentService) buildSolanaCollectionIntentRequest(request dto.SolanaCreateIntentDTO) (CollectionIntentRequestPayload, error) {
	if s.provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("solana collection intent requires Solana provider")
	}
	if request.WalletAddress == "" {
		return nil, fmt.Errorf("wallet address is required for solana collection intent")
	}
	if request.PledgeID == "" {
		return nil, fmt.Errorf("pledge id is required for solana collection intent")
	}
	if request.Network == "" {
		return nil, fmt.Errorf("network is required for solana collection intent")
	}
	return SolanaCollectionIntentRequest{
		Amount:       request.AmountLamports,
		Currency:     "usd",
		PayerPubkey:  request.WalletAddress,
		PledgeID:     request.PledgeID,
		DeadlineUnix: request.DeadlineUnix,
		Network:      request.Network,
		ProgramID:    request.ProgramID,
		OraclePubkey: request.OraclePubkey,
	}, nil
}

func (s *PaymentService) buildSolanaResolvePledgeRequest(request dto.SolanaResolvePledgeDTO) (SettlementResolutionRequestPayload, error) {
	if s.provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("solana pledge resolution requires Solana provider")
	}
	// DTO-level validation (constructors) should ensure required fields.
	// Normalize resolution for adapter request.
	resolution := strings.ToLower(strings.TrimSpace(request.Resolution))

	return SolanaSettlementResolutionRequest{
		ProviderReference: request.Operation,
		Resolution:        resolution,
		PledgePDA:         request.PledgePDA,
		UserPubkey:        request.UserPubkey,
		PenaltyPoolKey:    request.PenaltyPoolKey,
		TxHashProof:       request.TxHashProof,
		Network:           request.Network,
		Operation:         request.Operation,
	}, nil
}

func extractClientSecretFromCollectionIntent(payload CollectionIntentResultPayload) (string, error) {
	switch v := payload.(type) {
	case *StripeCollectionIntentResult:
		return v.ClientSecret, nil
	case *SolanaCollectionIntentResult:
		return v.ClientSecret, nil
	default:
		return "", fmt.Errorf("unsupported collection intent result payload type %T", payload)
	}
}

func extractReferenceAndStatusFromSettlementIntent(payload SettlementIntentResultPayload) (string, entities.SettlementStatus, error) {
	switch v := payload.(type) {
	case *StripeSettlementIntentResult:
		return v.ProviderReference, v.Status, nil
	case *SolanaSettlementIntentResult:
		return v.ProviderReference, v.Status, nil
	default:
		return "", "", fmt.Errorf("unsupported settlement intent result payload type %T", payload)
	}
}

func extractReferenceAndStatusFromDisbursement(payload DisbursementResultPayload) (string, entities.SettlementStatus, error) {
	switch v := payload.(type) {
	case *StripeDisbursementResult:
		return v.ProviderReference, v.Status, nil
	case *SolanaDisbursementResult:
		return v.ProviderReference, v.Status, nil
	default:
		return "", "", fmt.Errorf("unsupported disbursement result payload type %T", payload)
	}
}

func buildPaymentSettlementDTO(settlement *entities.PaymentSettlement) dto.PaymentSettlementDTO {
	if settlement == nil {
		return dto.PaymentSettlementDTO{}
	}

	return dto.PaymentSettlementDTO{
		ID:              settlement.ID,
		UserID:          settlement.UserID,
		Operation:       settlement.Operation,
		IdempotencyKey:  settlement.IdempotencyKey,
		Provider:        settlement.Provider,
		PaymentMethodID: settlement.PaymentMethodID,
		Status:          settlement.Status,
		Amount:          settlement.Amount,
		Currency:        settlement.Currency,
		RetryCount:      settlement.RetryCount,
		LastError:       settlement.LastError,
		Reference: dto.SettlementReferenceDTO{
			ProviderReference: settlement.Reference.ProviderReference,
			Network:           settlement.Reference.Network,
			TxHash:            settlement.Reference.TxHash,
			ContractAddress:   settlement.Reference.ContractAddress,
			SettlementProof:   settlement.Reference.SettlementProof,
			FinalizedAtUnix:   settlement.Reference.FinalizedAtUnix,
		},
		CreatedAtUnix: settlement.CreatedAt.Unix(),
		UpdatedAtUnix: settlement.UpdatedAt.Unix(),
	}
}

// ------------------------------------------------------------
// Stripe-specific PaymentService methods
// Methods in this section rely on Stripe provider behavior.
// ------------------------------------------------------------

func (s *PaymentService) CreateStripeCollectionIntent(request dto.StripeCreateIntentDTO, idempotencyKey string) (*dto.StripeCreateCollectionIntentResultDTO, error) {

	// use a provider-scoped operation name for idempotency to avoid collision
	claimed, stored, err := s.claimOrGetStoredResponse("stripe_collection_intent", idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !claimed {
		if stored != "" {
			return &dto.StripeCreateCollectionIntentResultDTO{ClientSecret: stored, IdempotentReplay: true}, nil
		}
		return &dto.StripeCreateCollectionIntentResultDTO{ClientSecret: "", IdempotentReplay: true}, nil
	}

	intentReq, reqErr := s.buildStripeCollectionIntentRequest(request)
	if reqErr != nil {
		return nil, reqErr
	}

	intentPayload, err := s.adapter.CreateCollectionIntent(intentReq)
	if err != nil {
		return nil, err
	}

	clientSecret, err := extractClientSecretFromCollectionIntent(intentPayload)
	if err != nil {
		return nil, err
	}

	s.setStoredResponse("stripe_collection_intent", idempotencyKey, clientSecret)

	return &dto.StripeCreateCollectionIntentResultDTO{ClientSecret: clientSecret, IdempotentReplay: false}, nil
}

func (s *PaymentService) addStripeCardMethod(request dto.AddPaymentMethodDTO) (*dto.AddPaymentMethodResultDTO, error) {
	if s.provider != entities.PaymentProviderStripe {
		return nil, fmt.Errorf("card payment method requires Stripe provider")
	}
	if s.cardAdapter == nil {
		return nil, fmt.Errorf("card onboarding is not available for provider %s", s.provider)
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
	return &dto.AddPaymentMethodResultDTO{PaymentMethod: *paymentMethodInfo}, nil
}

// ------------------------------------------------------------
// Solana-specific PaymentService methods (non-custodial signing flow)
// These methods implement the sign-then-submit flow for Solana.
// ------------------------------------------------------------

func (s *PaymentService) addSolanaWalletMethod(request dto.AddPaymentMethodDTO) (*dto.AddPaymentMethodResultDTO, error) {
	if s.provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("solana wallet requires Solana provider")
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
	return &dto.AddPaymentMethodResultDTO{PaymentMethod: *paymentMethodInfo}, nil
}

func (s *PaymentService) CreateSolanaCollectionIntent(request dto.SolanaCreateIntentDTO, idempotencyKey string) (*dto.SolanaCreateCollectionIntentResultDTO, error) {
	if s.provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("solana collection intent requires Solana provider")
	}

	claimed, stored, err := s.claimOrGetStoredResponse("solana_collection_intent", idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !claimed {
		if stored != "" {
			var resp dto.SolanaCreateCollectionIntentResultDTO
			if unmarshalErr := json.Unmarshal([]byte(stored), &resp); unmarshalErr == nil {
				return &resp, nil
			}
		}
		if replay, replayErr := s.replaySolanaCollectionIntent(idempotencyKey); replayErr == nil {
			return replay, nil
		}
	}

	intentReq, reqErr := s.buildSolanaCollectionIntentRequest(request)
	if reqErr != nil {
		return nil, reqErr
	}

	intentPayload, err := s.adapter.CreateCollectionIntent(intentReq)
	if err != nil {
		return nil, err
	}

	result, ok := intentPayload.(*SolanaCollectionIntentResult)
	if !ok {
		return nil, fmt.Errorf("unexpected solana collection intent result type %T", intentPayload)
	}

	if s.settlementRepo != nil {
		settlement := entities.NewPaymentSettlement(request.UserID, "solana_collection_intent", idempotencyKey, s.provider, request.WalletAddress, request.AmountLamports)
		settlement.Status = entities.SettlementStatusPending
		settlement.Reference.ProviderReference = result.ProviderReference
		settlement.Reference.Network = request.Network
		settlement.Reference.ContractAddress = request.ProgramID
		settlement.Reference.FinalizedAtUnix = 0
		_, _ = s.settlementRepo.Create(settlement)
	}

	response := &dto.SolanaCreateCollectionIntentResultDTO{
		ProviderReference: result.ProviderReference,
		UnsignedTxJSON:    result.UnsignedTxJSON,
		PledgePDA:         result.PledgePDA,
		RecentBlockhash:   result.RecentBlockhash.String(),
		ExpiresAtUnix:     result.ExpiresAtUnix,
	}

	if encoded, marshalErr := json.Marshal(response); marshalErr == nil {
		s.setStoredResponse("solana_collection_intent", idempotencyKey, string(encoded))
	}

	return response, nil
}

func (s *PaymentService) SubmitSolanaSignedTransaction(request dto.SolanaSubmitSignedTransactionDTO, idempotencyKey string) (*dto.SolanaSubmitSignedTransactionResultDTO, error) {
	if s.provider != entities.PaymentProviderSolana {
		return nil, fmt.Errorf("solana signed transaction submission requires Solana provider")
	}

	claimed, stored, err := s.claimOrGetStoredResponse("solana_submit_signed_transaction", idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !claimed {
		if stored != "" {
			var resp dto.SolanaSubmitSignedTransactionResultDTO
			if unmarshalErr := json.Unmarshal([]byte(stored), &resp); unmarshalErr == nil {
				return &resp, nil
			}
		}
		if replay, replayErr := s.replaySolanaSubmitResult(idempotencyKey); replayErr == nil {
			return replay, nil
		}
	}

	rpcProvider, ok := any(s.adapter).(solanaConfigProvider)
	if !ok || rpcProvider.RPCEndpoint() == "" {
		return nil, fmt.Errorf("solana adapter does not expose an RPC endpoint")
	}

	signedTxBytes, decodeErr := base64.StdEncoding.DecodeString(request.SignedTransactionBase64)
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid signed transaction base64: %w", decodeErr)
	}

	if s.settlementRepo == nil {
		return nil, fmt.Errorf("solana settlement repository is required to submit signed transactions")
	}

	settlement, findErr := s.settlementRepo.FindByOperationAndKey("solana_collection_intent", idempotencyKey)
	if findErr != nil {
		return nil, fmt.Errorf("solana settlement not found for idempotency key %s: %w", idempotencyKey, findErr)
	}
	if settlement.Reference.ProviderReference != "" && settlement.Reference.ProviderReference != request.ProviderReference {
		return nil, fmt.Errorf("provider reference mismatch: expected %s, got %s", settlement.Reference.ProviderReference, request.ProviderReference)
	}

	submittedSig, submitErr := controlsolana.SubmitTransactionWithRetry(
		rpcProvider.RPCEndpoint(),
		controlsolana.SignedTransaction{Bytes: signedTxBytes},
		3,
	)
	if submitErr != nil {
		settlement.Status = entities.SettlementStatusFailed
		settlement.LastError = submitErr.Error()
		_, _ = s.settlementRepo.Update(settlement)
		return nil, submitErr
	}

	signature := solana.SignatureFromBytes(submittedSig[:]).String()
	proof := map[string]any{
		"provider_reference": request.ProviderReference,
		"signature":          signature,
		"signed_tx_base64":   request.SignedTransactionBase64,
		"submitted_at_unix":  time.Now().Unix(),
		"network":            request.Network,
	}
	proofJSON, _ := json.Marshal(proof)

	settlement.Status = entities.SettlementStatusSettledOnChain
	settlement.Reference.ProviderReference = request.ProviderReference
	settlement.Reference.TxHash = signature
	settlement.Reference.SettlementProof = string(proofJSON)
	settlement.Reference.FinalizedAtUnix = time.Now().Unix()
	settlement.LastError = ""
	_, _ = s.settlementRepo.Update(settlement)

	response := &dto.SolanaSubmitSignedTransactionResultDTO{
		ProviderReference: request.ProviderReference,
		Signature:         signature,
		Status:            string(settlement.Status),
		SettlementProof:   string(proofJSON),
	}

	if encoded, marshalErr := json.Marshal(response); marshalErr == nil {
		s.setStoredResponse("solana_submit_signed_transaction", idempotencyKey, string(encoded))
	}

	return response, nil
}

func (s *PaymentService) replaySolanaCollectionIntent(idempotencyKey string) (*dto.SolanaCreateCollectionIntentResultDTO, error) {
	if s.idempotencyRepo != nil {
		stored, err := s.idempotencyRepo.GetResponse("solana_collection_intent", idempotencyKey)
		if err == nil && stored != "" {
			var response dto.SolanaCreateCollectionIntentResultDTO
			if unmarshalErr := json.Unmarshal([]byte(stored), &response); unmarshalErr == nil {
				return &response, nil
			}
		}
	}

	if s.settlementRepo == nil {
		return nil, fmt.Errorf("solana collection intent replay unavailable")
	}

	settlement, err := s.settlementRepo.FindByOperationAndKey("solana_collection_intent", idempotencyKey)
	if err != nil {
		return nil, err
	}

	return &dto.SolanaCreateCollectionIntentResultDTO{
		ProviderReference: settlement.Reference.ProviderReference,
		UnsignedTxJSON:    "",
		PledgePDA:         settlement.Reference.ContractAddress,
		RecentBlockhash:   "",
		ExpiresAtUnix:     settlement.Reference.FinalizedAtUnix,
	}, nil
}

func (s *PaymentService) replaySolanaSubmitResult(idempotencyKey string) (*dto.SolanaSubmitSignedTransactionResultDTO, error) {
	if s.idempotencyRepo != nil {
		stored, err := s.idempotencyRepo.GetResponse("solana_submit_signed_transaction", idempotencyKey)
		if err == nil && stored != "" {
			var response dto.SolanaSubmitSignedTransactionResultDTO
			if unmarshalErr := json.Unmarshal([]byte(stored), &response); unmarshalErr == nil {
				return &response, nil
			}
		}
	}

	if s.settlementRepo == nil {
		return nil, fmt.Errorf("solana submit replay unavailable")
	}

	settlement, err := s.settlementRepo.FindByOperationAndKey("solana_collection_intent", idempotencyKey)
	if err != nil {
		return nil, err
	}

	return &dto.SolanaSubmitSignedTransactionResultDTO{
		ProviderReference: settlement.Reference.ProviderReference,
		Signature:         settlement.Reference.TxHash,
		Status:            string(settlement.Status),
		SettlementProof:   settlement.Reference.SettlementProof,
	}, nil
}
