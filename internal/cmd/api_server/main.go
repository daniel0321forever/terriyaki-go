package main

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/daniel0321forever/terriyaki-go/internal/interface/api"
	"github.com/daniel0321forever/terriyaki-go/internal/migrate"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// configure CORS middleware
	corsConfig := config.ConfigureCORS()
	router.Use(cors.New(corsConfig))

	// connect to database
	db, err := postgres.Connect()
	if err != nil {
		panic(err)
	}

	err = migrate.MigrateDatabase()
	if err != nil {
		panic(err)
	}

	// Initialize repositories
	userRepo := postgres.NewGormUserRepository(db)
	grindRepo := postgres.NewGormGrindRepository(db)
	taskRepo := postgres.NewGormTaskRepository(db)
	participationRepo := postgres.NewGormParticipationRepository(db)
	messageRepo := postgres.NewGormMessageRepository(db)
	interviewSessionRepo := postgres.NewGormInterviewSessionRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo)
	grindService := services.NewGrindService(grindRepo, userRepo, taskRepo, participationRepo, messageRepo)
	taskService := services.NewTaskService(taskRepo)
	messageService := services.NewMessageService(messageRepo, userRepo)
	interviewService := services.NewInterviewService(interviewSessionRepo)

	// Initialize API handlers with services
	grindCtrl := api.NewGrindController(grindService, userService, messageService)
	userCtrl := api.NewUserController(grindService, userService)
	healthCtrl := api.NewHealthController()
	taskCtrl := api.NewTaskController(taskService, grindService)
	messageCtrl := api.NewMessageController(userService, messageService, grindService)
	interviewCtrl := api.NewInterviewController(interviewService, userService, taskService)

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
		v1.POST("tasks/finish", taskCtrl.FinishTodayTaskAPI)
		v1.GET("tasks/today", taskCtrl.GetTodayTaskAPI)
		v1.GET("tasks/:id", taskCtrl.GetTaskAPI)
		v1.GET("messages", messageCtrl.GetMessageAPI)
		v1.POST("messages/:id/read", messageCtrl.ReadMessageAPI)
		v1.POST("messages/:id/invitation/create", messageCtrl.CreateInvitationAPI)
		v1.POST("messages/:id/invitation/accept", messageCtrl.AcceptInvitationAPI)
		v1.POST("messages/:id/invitation/reject", messageCtrl.RejectInvitationAPI)
		v1.POST("interviews/llm", interviewCtrl.LLMWebhookAPI)
		v1.POST("interviews/start", interviewCtrl.StartInterviewAPI)
		v1.POST("interviews/:id/response", interviewCtrl.SaveAgentResponseAPI)
		v1.POST("interviews/:id/end", interviewCtrl.EndInterviewAPI)
	}

	router.Run(":8080")
}
