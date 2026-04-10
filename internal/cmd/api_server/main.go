package main

import (
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/container"
	"github.com/daniel0321forever/terriyaki-go/migrations"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/daniel0321forever/terriyaki-go/internal/interface/api"
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

	err = migrations.MigrateDatabase(db)
	if err != nil {
		panic(err)
	}

	// NOTE: with this we can easily change the repository implementation without changing the codebase
	err = container.InitializeReposContainer(db)
	if err != nil {
		panic(err)
	}

	api.RegisterRoutes(router, db)

	router.Run(":8080")
}
