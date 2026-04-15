package services

import (
	"fmt"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"github.com/redis/go-redis/v9"
)

// Payment service interface
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
		@param request - the save card request
		@return the card id
	*/
	SaveCard(request dto.SaveCardDTO) (*entities.PaymentMethodInfo, error)

	/*
		Charge a payment intent
		@param amount - the amount to charge
		@return the payment intent id
	*/
	Charge(paymentInfo entities.PaymentMethodInfo, amount int64) (string, error)

	/*
		Pay back a payment intent
		@param userStripeAccountID - the destination account id
		@param amount - the amount to pay back
		@return the payment intent id
	*/
	PayBack(userStripeAccountID string, amount int64) error

	/*
		Find dued payments
		@param userID - the user id
		@return the dued payments
	*/
	FindDuedPayments() ([]dto.PendingPaymentDTO, error)

	/*
		Get available payment methods for user
		@param request - the query request
		@return available payment methods
	*/
	GetAvailablePaymentMethods(request dto.GetAvailablePaymentMethodsDTO) (*dto.AvailablePaymentMethodsDTO, error)
}

type PaymentService struct {
	adapter PaymentGatewayAdapter

	/*
		The Redis client
		@type *redis.Client
	*/
	RedisClient *redis.Client

	/*
		The user service
		@type *UserService
	*/
	userRepo repositories.UserRepository

	/*
		The grind service
		@type *GrindService
	*/
	grindRepo repositories.GrindRepository

	/*
		The participation repository
		@type *ParticipationRepository
	*/
	participationRepo repositories.ParticipationRepository

	/*
		The stripe payment info repository
		@type *StripePaymentInfoRepository
	*/
	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository
}

func NewPaymentService(
	userRepo repositories.UserRepository,
	grindRepo repositories.GrindRepository,
	participationRepo repositories.ParticipationRepository,
	paymentMethodInfoRepo repositories.PaymentMethodInfoRepository,
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
	}
}

func (s *PaymentService) attachPaymentMethodToCustomer(stripePaymentMethodID string, stripeCustomerID string) error {
	return s.adapter.AttachPaymentMethodToCustomer(stripePaymentMethodID, stripeCustomerID)
}

func (s *PaymentService) CreatePaymentIntent(amount int64) (string, error) {
	return s.adapter.CreatePaymentIntent(amount)
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

func (s *PaymentService) Charge(paymentInfo entities.PaymentMethodInfo, amount int64) (string, error) {
	return s.adapter.Charge(paymentInfo.ProviderCustomerID, paymentInfo.ProviderPaymentMethodID, amount)
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
