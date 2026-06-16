package api

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
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
	grindService := services.NewGrindService(grindRepo, userRepo, habitTaskRepo, participationRepo, messageRepo)
	messageService := services.NewMessageService(messageRepo, userRepo, grindRepo)
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
	healthCtrl := NewHealthController()
	messageCtrl := NewMessageController(userService, messageService, grindService)
	paymentCtrl := NewPaymentController(userService, stripePaymentService, solanaPaymentService)
	profileCtrl := NewProfileController(userService)
	ingestCtrl := NewIngestController(ingestService)
	partnerGroupCtrl := NewPartnerGroupController(partnerGroupService)

	// define routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("grinds", grindCtrl.CreateGrindAPI)
		v1.GET("grinds", grindCtrl.GetAllUserGrindsAPI)
		v1.DELETE("grinds/delete-all", grindCtrl.DeleteAllGrindsAPI)
		v1.GET("grinds/current", grindCtrl.GetUserCurrentGrindAPI)
		v1.GET("grinds/:id", grindCtrl.GetGrindAPI)
		v1.GET("ping", healthCtrl.PingAPI)
		v1.POST("register", userCtrl.RegisterAPI)
		v1.POST("login", userCtrl.LoginAPI)
		v1.POST("logout", userCtrl.LogoutAPI)
		v1.GET("verify-token", userCtrl.VerifyTokenAPI)
		// Stripe payment endpoints
		v1.POST("payments/stripe/payment-intent", paymentCtrl.PaymentIntentAPI)
		v1.POST("payments/methods", paymentCtrl.AddPaymentMethodAPI)
		v1.POST("payments/stripe/force-charging", paymentCtrl.ForceInvestigateDuedPenaltyAPI)
		v1.GET("payments/stripe/methods", paymentCtrl.GetAvailablePaymentMethodsAPI)
		v1.POST("payments/stripe/methods/select-default", paymentCtrl.SelectPaymentMethodAPI)
		// Solana payment endpoints
		v1.POST("payments/solana/collection-intent", paymentCtrl.CreateSolanaCollectionIntentAPI)
		v1.POST("payments/solana/submit-signed-transaction", paymentCtrl.SubmitSolanaSignedTransactionAPI)
		v1.GET("users/exists", userCtrl.CheckUserExistsAPI)
		v1.PATCH("users/update-profile", profileCtrl.UpdateProfileAPI)
		v1.GET("messages", messageCtrl.GetMessageAPI)
		v1.POST("messages/invitation", messageCtrl.CreateInvitationAPI)
		v1.POST("messages/:id/invitation/accept", messageCtrl.AcceptInvitationAPI)
		v1.POST("messages/:id/invitation/reject", messageCtrl.RejectInvitationAPI)
		v1.GET("messages/sent", messageCtrl.GetSentMessageAPI)
		v1.POST("messages/:id/read", messageCtrl.ReadMessageAPI)
	}

	v2 := router.Group("/api/v2")
	{
		v2.POST("login", userCtrl.LoginAPIV2)
		v2.GET("verify-token", userCtrl.VerifyTokenAPIV2)
		v2.POST("grinds/:id/quit", grindCtrl.QuitGrindAPI)
		v2.POST("ingest/:provider", ingestCtrl.HandleIngest)
		v2.POST("groups", partnerGroupCtrl.CreateGroupAPI)
		// Register static POST groups/join BEFORE dynamic GET groups/:id to avoid Gin wildcard conflict.
		v2.POST("groups/join", partnerGroupCtrl.JoinGroupAPI)
		v2.GET("groups/:id", partnerGroupCtrl.GetGroupAPI)
		v2.POST("groups/:id/invite", partnerGroupCtrl.GenerateInviteLinkAPI)
	}
}
