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
	router.POST("/v1/register", api.RegisterAPI)
	router.POST("/v1/login", api.LoginAPI)
	router.POST("/v1/logout", api.LogoutAPI)
	router.GET("/v1/verify-token", api.VerifyTokenAPI)
	router.DELETE("/v1/users/delete", api.DeleteUserAPI)
	router.POST("/v1/grinds", api.CreateGrindAPI)
	router.GET("/v1/grinds", api.GetAllUserGrindsAPI)
	router.DELETE("/v1/grinds/delete-all", api.DeleteAllGrindsAPI)
	router.GET("/v1/grinds/current", api.GetUserCurrentGrindAPI)
	router.GET("/v1/grinds/:id", api.GetGrindAPI)
	router.POST("/v1/grinds/:id/quit", api.QuitGrindAPI)
	router.POST("/v1/tasks/finish", api.FinishTodayTaskAPI)
	router.GET("/v1/tasks/:id", api.GetTaskAPI)
	router.GET("/v1/messages", api.GetMessageAPI)
	router.POST("/v1/messages/:id/read", api.ReadMessageAPI)
	router.POST("/v1/messages/:id/invitation/create", api.CreateInvitationAPI)
	router.POST("/v1/messages/:id/invitation/accept", api.AcceptInvitationAPI)
	router.POST("/v1/messages/:id/invitation/reject", api.RejectInvitationAPI)

	router.Run(":8080")
}
