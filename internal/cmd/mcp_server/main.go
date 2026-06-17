// Package main is the entry point for the Habitat MCP server binary.
// It initialises PostgreSQL and Redis, constructs the minimal repository and
// service graph needed by the two MCP tools, registers the tools, and then
// calls mcpserver.ServeStdio to block on stdin/stdout.
//
// Launch pattern (Claude CLI / Claude Desktop):
//
//	{
//	  "mcpServers": {
//	    "habitat": { "command": "/path/to/habitat-mcp" }
//	  }
//	}
//
// The binary reads the same environment variables as the API server:
// POSTGRES_DSN (or equivalent consumed by postgres.Connect), REDIS_ADDR,
// REDIS_PASSWORD.
package main

import (
	"os"

	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	mcpmux "github.com/daniel0321forever/terriyaki-go/internal/interface/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Connect to PostgreSQL — same call as api_server/main.go.
	db, err := postgres.Connect()
	if err != nil {
		panic(err)
	}

	// Initialise Redis client. Credentials come from environment variables.
	// The connection pool is intentionally not closed (lives for process lifetime).
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv(config.REDIS_ADDR),
		Password: os.Getenv(config.REDIS_PASSWORD),
		DB:       0,
		Protocol: 2,
	})
	_ = rdb // retained for future MCP tools that need Redis; not used by current handlers

	// Construct only the repositories required by the two MCP tools.
	habitTaskRepo := postgres.NewGormHabitTaskRepository(db)
	completionEventRepo := postgres.NewGormCompletionEventRepository(db)

	// Build the application service.
	ingestService := services.NewIngestService(habitTaskRepo, completionEventRepo)

	// Register tools with the MCP server.
	s := mcpserver.NewMCPServer("habitat-mcp", "0.1.0")
	s.AddTool(mcpmux.IngestCompletionEventTool, mcpmux.HandleIngestCompletionEvent(ingestService))
	s.AddTool(mcpmux.VerifyHabitTaskTool, mcpmux.HandleVerifyHabitTask(habitTaskRepo, completionEventRepo))

	// Serve on stdio transport — blocks until the process exits.
	if err := mcpserver.ServeStdio(s); err != nil {
		panic(err)
	}
}
