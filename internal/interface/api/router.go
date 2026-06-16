package api

import (
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/daniel0321forever/terriyaki-go/internal/interface/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	// Initialize repositories
	userRepo := postgres.NewGormUserRepository(db)
	grindRepo := postgres.NewGormGrindRepository(db)
	participationRepo := postgres.NewGormParticipationRepository(db)
	messageRepo := postgres.NewGormMessageRepository(db)
	paymentInfoRepo := postgres.NewGormStripePaymentInfoRepository(db)
	paymentIdempotencyRepo := postgres.NewGormPaymentIdempotencyRepository(db)
	paymentSettlementRepo := postgres.NewGormPaymentSettlementRepository(db)
	habitTaskRepo := postgres.NewGormHabitTaskRepository(db)
	completionEventRepo := postgres.NewGormCompletionEventRepository(db)
	partnerGroupRepo := postgres.NewGormPartnerGroupRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo)
	grindService := services.NewGrindService(db, grindRepo, userRepo, habitTaskRepo, participationRepo, messageRepo)
	messageService := services.NewMessageService(db, messageRepo, userRepo, grindRepo)
	ingestService := services.NewIngestService(habitTaskRepo, completionEventRepo)
	partnerGroupService := services.NewPartnerGroupService(partnerGroupRepo)
	paymentFactory := services.NewPaymentServiceFactory(
		userRepo,
		grindRepo,
		participationRepo,
		paymentInfoRepo,
		paymentIdempotencyRepo,
		paymentSettlementRepo,
	)

	// Initialize payment service bound to a single provider (Stripe).
	// The provider is selected at startup via the factory.
	stripePaymentService, err := paymentFactory.BuildForProvider(
		entities.PaymentProviderStripe,
	)
	if err != nil {
		panic(err)
	}
	solanaPaymentService, err := paymentFactory.BuildForProvider(
		entities.PaymentProviderSolana,
	)
	if err != nil {
		panic(err)
	}

	// Initialize API handlers with services
	grindCtrl := NewGrindController(grindService, userService, messageService)
	userCtrl := NewUserController(grindService, userService)
	healthCtrl := NewHealthController(db, rdb)
	messageCtrl := NewMessageController(userService, messageService, grindService)
	paymentCtrl := NewPaymentController(userService, stripePaymentService, solanaPaymentService)
	profileCtrl := NewProfileController(userService)
	ingestCtrl := NewIngestController(ingestService)
	partnerGroupCtrl := NewPartnerGroupController(partnerGroupService)

	// Rate limit middleware: 10 requests per minute per IP (SEC-03)
	// Fail-open: Redis error allows request through (T-03-06 mitigated).
	rl := middleware.RateLimitMiddleware(rdb, 10, time.Minute)

	// Health check on root router (unversioned, per D-04)
	router.GET("/api/health", healthCtrl.HealthAPI)

	v2 := router.Group("/api/v2")
	{
		// Auth routes (v2 handlers win per D-02) — rate limited (T-03-05)
		v2.POST("login", rl, userCtrl.LoginAPIV2)
		v2.GET("verify-token", userCtrl.VerifyTokenAPIV2)

		// Register static grind paths BEFORE dynamic :id
		v2.POST("grinds", grindCtrl.CreateGrindAPI)
		v2.GET("grinds", grindCtrl.GetAllUserGrindsAPI)
		v2.GET("grinds/current", grindCtrl.GetUserCurrentGrindAPI) // static BEFORE grinds/:id
		v2.GET("grinds/:id", grindCtrl.GetGrindAPI)
		v2.POST("grinds/:id/quit", grindCtrl.QuitGrindAPI)

		// User routes — register rate limited (T-03-05)
		v2.POST("register", rl, userCtrl.RegisterAPI)
		v2.POST("logout", userCtrl.LogoutAPI)
		v2.GET("users/exists", userCtrl.CheckUserExistsAPI)
		v2.PATCH("users/update-profile", profileCtrl.UpdateProfileAPI)

		// Payment routes (Stripe)
		v2.POST("payments/stripe/payment-intent", paymentCtrl.PaymentIntentAPI)
		v2.POST("payments/methods", paymentCtrl.AddPaymentMethodAPI)
		v2.POST("payments/stripe/force-charging", paymentCtrl.ForceInvestigateDuedPenaltyAPI)
		v2.GET("payments/stripe/methods", paymentCtrl.GetAvailablePaymentMethodsAPI)
		v2.POST("payments/stripe/methods/select-default", paymentCtrl.SelectPaymentMethodAPI)

		// Payment routes (Solana)
		v2.POST("payments/solana/collection-intent", paymentCtrl.CreateSolanaCollectionIntentAPI)
		v2.POST("payments/solana/submit-signed-transaction", paymentCtrl.SubmitSolanaSignedTransactionAPI)

		// Message routes — register static paths BEFORE dynamic :id
		v2.GET("messages", messageCtrl.GetMessageAPI)
		v2.POST("messages/invitation", messageCtrl.CreateInvitationAPI) // static BEFORE messages/:id
		v2.GET("messages/sent", messageCtrl.GetSentMessageAPI)          // static BEFORE messages/:id
		v2.POST("messages/:id/invitation/accept", messageCtrl.AcceptInvitationAPI)
		v2.POST("messages/:id/invitation/reject", messageCtrl.RejectInvitationAPI)
		v2.POST("messages/:id/read", messageCtrl.ReadMessageAPI)

		// Ingest — rate limited (T-03-05)
		v2.POST("ingest/:provider", rl, ingestCtrl.HandleIngest)

		// Partner groups — register static POST groups/join BEFORE dynamic GET groups/:id
		v2.POST("groups", partnerGroupCtrl.CreateGroupAPI)
		v2.POST("groups/join", partnerGroupCtrl.JoinGroupAPI)
		v2.GET("groups/:id", partnerGroupCtrl.GetGroupAPI)
		v2.POST("groups/:id/invite", partnerGroupCtrl.GenerateInviteLinkAPI)
	}
}
