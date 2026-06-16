package main

import (
	"os"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/container"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/daniel0321forever/terriyaki-go/internal/interface/api"
	"github.com/daniel0321forever/terriyaki-go/migrations"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

	// Initialize Redis client. Credentials come from environment (T-03-08: never hardcode).
	// Do NOT close rdb in a defer — the connection pool lives for the full process lifetime.
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv(config.REDIS_ADDR),
		Password: os.Getenv(config.REDIS_PASSWORD),
		DB:       0,
		Protocol: 2,
	})

	api.RegisterRoutes(router, db, rdb)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
