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
	CreateSaveCardIntent() (string, error)
	SaveCard(request dto.SaveCardDTO) (*entities.PaymentMethodInfo, error)

	// Charging lifecycle
	ChargeWithIdempotency(paymentInfo entities.PaymentMethodInfo, amount int64, operation string, idempotencyKey string, userID string) (string, bool, error)

	PayBack(userStripeAccountID string, amount int64) error

	// Settlement/reconciliation
	FindDuedPayments() ([]dto.PendingPaymentDTO, error)
	ReconcileSettlements(limit int) ([]entities.PaymentSettlement, error)

	// Queries and idempotency helpers
	GetAvailablePaymentMethods(request dto.GetAvailablePaymentMethodsDTO) (*dto.AvailablePaymentMethodsDTO, error)
	ClaimIdempotency(operation string, idempotencyKey string) (bool, error)
}

type PaymentService struct {
	adapter PaymentGatewayAdapter
	// RedisClient is currently kept for future queue/retry usage.
	RedisClient *redis.Client

	userRepo repositories.UserRepository
	grindRepo repositories.GrindRepository
	participationRepo repositories.ParticipationRepository

	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository
	idempotencyRepo      repositories.PaymentIdempotencyRepository
	settlementRepo       repositories.PaymentSettlementRepository
}

func NewPaymentService(
	userRepo repositories.UserRepository,
	grindRepo repositories.GrindRepository,
	participationRepo repositories.ParticipationRepository,
	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository,
	idempotencyRepo repositories.PaymentIdempotencyRepository,
	settlementRepo repositories.PaymentSettlementRepository,
	adapter PaymentGatewayAdapter,
) *PaymentService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})

	if adapter == nil {
		return nil
	}

	return &PaymentService{
		adapter:               adapter,
		RedisClient:           rdb,
		userRepo:              userRepo,
		grindRepo:             grindRepo,
		participationRepo:     participationRepo,
		paymentMethodInfoRepo: paymentMethodInfoRepo,
		idempotencyRepo:       idempotencyRepo,
		settlementRepo:        settlementRepo,
	}
}

func (s *PaymentService) attachPaymentMethodToCustomer(stripePaymentMethodID string, stripeCustomerID string) error {
	return s.adapter.AttachPaymentMethodToCustomer(stripePaymentMethodID, stripeCustomerID)
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

	clientSecret, err := s.adapter.CreatePaymentIntent(amount)
	if err != nil {
		return "", false, err
	}

	if s.idempotencyRepo != nil {
		_ = s.idempotencyRepo.SetResponse("payment_intent", idempotencyKey, clientSecret)
	}

	return clientSecret, false, nil
}

func (s *PaymentService) CreateSaveCardIntent() (string, error) {
	return s.adapter.CreateSaveCardIntent()
}

func (s *PaymentService) SaveCard(
	request dto.SaveCardDTO,
) (*entities.PaymentMethodInfo, error) {

	// create stripe customer if not exists
	var stripeCustomerID string

	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return nil, err
	}

	if user.StripeCustomerID != "" {
		stripeCustomerID = user.StripeCustomerID
	} else {
		createdCustomerID, err := s.adapter.CreateCustomer(user.Username, user.Email)
		if err != nil {
			return nil, err
		}

		fmt.Print("created customer id " + createdCustomerID + " for user " + user.Username)
		stripeCustomerID = createdCustomerID

		user.StripeCustomerID = stripeCustomerID
		user.DefaultPaymentMethodID = request.PaymentMethodID
		err = s.userRepo.Update(user)
		if err != nil {
			return nil, err
		}
	}

	// attach the payment method to the customer
	err = s.attachPaymentMethodToCustomer(request.PaymentMethodID, stripeCustomerID)
	if err != nil {
		return nil, err
	}

	paymentMethodInfo, err := s.adapter.DescribePaymentMethod(request.PaymentMethodID)
	if err != nil {
		return nil, err
	}
	paymentMethodInfo.UserID = user.ID
	paymentMethodInfo.ProviderCustomerID = stripeCustomerID

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

	var settlement *entities.PaymentSettlement
	if s.settlementRepo != nil {
		settlement = entities.NewPaymentSettlement(userID, operation, idempotencyKey, paymentInfo.Provider, paymentInfo.ProviderPaymentMethodID, amount)
		settlement.Reference.Network = paymentInfo.Network
		settlement, err = s.settlementRepo.Create(settlement)
		if err != nil {
			return "", false, err
		}
	}

	reference, chargeErr := s.adapter.Charge(paymentInfo.ProviderCustomerID, paymentInfo.ProviderPaymentMethodID, amount)
	if chargeErr != nil {
		if settlement != nil {
			settlement.Status = entities.SettlementStatusFailed
			settlement.LastError = chargeErr.Error()
			_, _ = s.settlementRepo.Update(settlement)
		}
		return "", false, chargeErr
	}

	if settlement != nil {
		settlement.Status = entities.SettlementStatusCaptured
		settlement.Reference.ProviderReference = reference
		_, _ = s.settlementRepo.Update(settlement)
	}

	if s.idempotencyRepo != nil {
		_ = s.idempotencyRepo.SetResponse(operation, idempotencyKey, reference)
	}

	return reference, false, nil
}

func (s *PaymentService) PayBack(userStripeAccountID string, amount int64) error {
	return s.adapter.PayBack(userStripeAccountID, amount)
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
