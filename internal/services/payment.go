package services

import (
	"fmt"
	"os"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/types"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/customer"
	"github.com/stripe/stripe-go/v84/paymentintent"
	"github.com/stripe/stripe-go/v84/paymentmethod"
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
	SaveCard(paymentInfo models.StripePaymentInfo) error

	/*
		Charge a payment intent
		@param amount - the amount to charge
		@return the payment intent id
	*/
	Charge(paymentInfo models.StripePaymentInfo, amount int64) (string, error)

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
	FindDuedPayments() ([]types.PendingPayment, error)

	/*
		Select a payment method
		@param user - the user
		@param stripePaymentMethodID - the stripe payment method id
		@return the result of the selection
	*/
	SelectPaymentMethod(user *models.User, stripePaymentMethodID string) error
}

type StripePaymentService struct {
	/*
		The Stripe secret key
		@type string
	*/
	StripeSecretKey string

	/*
		The Redis client
		@type *redis.Client
	*/
	RedisClient *redis.Client
}

func (s *StripePaymentService) selectPaymentMethod(user *models.User, stripePaymentMethodID string) error {
	_, err := models.UpdateUser(user.ID, nil, nil, nil, nil, nil, &stripePaymentMethodID)
	if err != nil {
		return err
	}

	return nil
}

func (s *StripePaymentService) getDefaultPaymentMethod(user *models.User) (string, error) {
	return user.DefaultPaymentMethodID, nil
}

func (s *StripePaymentService) attachPaymentMethodToCustomer(stripePaymentMethodID string, stripeCustomerID string) error {
	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(stripeCustomerID),
	}
	paymentmethod.Attach(stripePaymentMethodID, attachParams)
	return nil
}

func NewStripePaymentService() (*StripePaymentService, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})

	stripeSecretKey := os.Getenv(config.OS_ENV_STRIPE_SECRET_KEY)
	if stripeSecretKey == "" {
		return nil, fmt.Errorf("%s", "environment variable"+config.OS_ENV_STRIPE_SECRET_KEY+"is not set")
	}

	return &StripePaymentService{
		StripeSecretKey: stripeSecretKey,
		RedisClient:     rdb,
	}, nil
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

func (s *StripePaymentService) SaveCard(
	user *models.User,
	stripePaymentMethodID string,
) (*models.StripePaymentInfo, error) {
	// get repo
	repo := &models.StripePaymentInfoRepository{}

	// create stripe customer if not exists
	var stripeCustomerID string
	if user.StripeCustomerID != "" {
		stripeCustomerID = user.StripeCustomerID
	} else {
		params := &stripe.CustomerParams{
			Name:  stripe.String(user.Username),
			Email: stripe.String(user.Email),
		}
		customer, err := customer.New(params)
		if err != nil {
			return nil, err
		}

		fmt.Print("created customer id " + customer.ID + " for user " + user.Username)
		stripeCustomerID = customer.ID

		_, err = models.UpdateUser(user.ID, nil, nil, nil, nil, &customer.ID, nil)
		if err != nil {
			return nil, err
		}

		err = s.selectPaymentMethod(user, stripePaymentMethodID)
		if err != nil {
			return nil, err
		}
	}

	// attach the payment method to the customer
	err := s.attachPaymentMethodToCustomer(stripePaymentMethodID, stripeCustomerID)
	if err != nil {
		return nil, err
	}

	// create instance
	instance, err := repo.Create(
		user.ID,
		stripeCustomerID,
		stripePaymentMethodID,
	)

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (s *StripePaymentService) Charge(paymentInfo models.StripePaymentInfo, amount int64) (string, error) {
	stripe.Key = s.StripeSecretKey

	pi, err := paymentintent.New(&stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      stripe.String(paymentInfo.StripeCustomerID),
		PaymentMethod: stripe.String(paymentInfo.StripePaymentMethodID),
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

func (s *StripePaymentService) SelectPaymentMethod(user *models.User, stripePaymentMethodID string) error {
	return s.selectPaymentMethod(user, stripePaymentMethodID)
}

func (s *StripePaymentService) GetAvailablePaymentMethods(user *models.User) ([]models.StripePaymentInfo, *models.StripePaymentInfo, error) {
	stripe.Key = s.StripeSecretKey

	// get payment infos
	repo := &models.StripePaymentInfoRepository{}
	paymentInfos, err := repo.GetByUserID(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get default payment method: %w", err)
	}

	// get default one
	defaultPaymentMethodID, err := s.getDefaultPaymentMethod(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get payment infos: %w", err)
	}

	var defaultPaymentInfo models.StripePaymentInfo
	for _, paymentInfo := range paymentInfos {
		if paymentInfo.StripePaymentMethodID == defaultPaymentMethodID {
			defaultPaymentInfo = paymentInfo
			break
		}
	}

	return paymentInfos, &defaultPaymentInfo, nil
}

func (s *StripePaymentService) FindDuedPayments() ([]types.PendingPayment, error) {
	var grinds []models.Grind
	// Find all grinds where StartDate + Duration (in days) is the current date (today, UTC)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	err := database.Db.
		Where("DATE(start_date, '+' || duration || ' day') = ?", today.Format("2006-01-02")).
		Find(&grinds).Error

	if err != nil {
		return nil, err
	}

	// get the punishment for each grind
	var pendingPayments []types.PendingPayment

	for _, g := range grinds {
		for _, p := range g.Participants {
			// get the stripe payment info for the user
			var stripePaymentInfo models.StripePaymentInfo
			err := database.Db.Where("user_id = ?", p.ID).First(&stripePaymentInfo).Error
			if err != nil {
				return nil, err
			}

			// get the total penalty for the user in the grind
			participateRecord, err := models.GetParticipateRecordByUserIDAndGrindID(p.ID, g.ID)
			if err != nil {
				return nil, err
			}

			pendingPayments = append(pendingPayments, types.PendingPayment{
				StripePaymentInfo: stripePaymentInfo,
				PaymentAmount:     int64(participateRecord.TotalPenalty),
			})
		}
	}

	return pendingPayments, nil
}
