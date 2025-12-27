package main

import (
	"github.com/daniel0321forever/terriyaki-go/api"
	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/migrate"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// configure CORS middleware
	config := config.ConfigureCORS()
	router.Use(cors.New(config))

	// connect to database
	_, err := database.Connect()
	if err != nil {
		panic(err)
	}

	err = migrate.MigrateDatabase()
	if err != nil {
		panic(err)
	}

	// define routes
	router.GET("/api/v1/ping", api.PingAPI)
	router.POST("/api/v1/register", api.RegisterAPI)
	router.POST("/api/v1/login", api.LoginAPI)
	router.POST("/api/v1/logout", api.LogoutAPI)
	router.GET("/api/v1/verify-token", api.VerifyTokenAPI)
	router.DELETE("/api/v1/users/delete", api.DeleteUserAPI)
	router.GET("/api/v1/users/exists", api.CheckUserExistsAPI)
	router.POST("/api/v1/grinds", api.CreateGrindAPI)
	router.GET("/api/v1/grinds", api.GetAllUserGrindsAPI)
	router.DELETE("/api/v1/grinds/delete-all", api.DeleteAllGrindsAPI)
	router.GET("/api/v1/grinds/current", api.GetUserCurrentGrindAPI)
	router.GET("/api/v1/grinds/:id", api.GetGrindAPI)
	router.POST("/api/v1/grinds/:id/quit", api.QuitGrindAPI)
	router.GET("/api/v1/grinds/:id/progress", api.GetProgressRecordsAPI)
	router.POST("/api/v1/tasks/finish", api.FinishTodayTaskAPI)
	router.GET("/api/v1/tasks/today", api.GetTodayTaskAPI)
	router.GET("/api/v1/tasks/:id", api.GetTaskAPI)
	router.GET("/api/v1/messages", api.GetMessageAPI)
	router.GET("/api/v1/messages/sent", api.GetSentMessageAPI)
	router.POST("/api/v1/messages/:id/read", api.ReadMessageAPI)
	router.POST("/api/v1/messages/invitation", api.CreateInvitationAPI)
	router.POST("/api/v1/messages/:id/invitation/accept", api.AcceptInvitationAPI)
	router.POST("/api/v1/messages/:id/invitation/reject", api.RejectInvitationAPI)
	router.POST("/api/v1/interviews/llm", api.LLMWebhookAPI)
	router.POST("/api/v1/interviews/start", api.StartInterviewAPI)
	router.POST("/api/v1/interviews/:id/response", api.SaveAgentResponseAPI)
	router.POST("/api/v1/interviews/:id/end", api.EndInterviewAPI)
	router.PATCH("/api/v1/profile", api.UpdateProfileAPI)

	router.POST("/api/v2/login", api.LoginAPIV2)
	router.GET("/api/v2/verify-token", api.VerifyTokenAPIV2)

	router.Run(":8080")
}
