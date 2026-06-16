// Package mcp provides MCP tool handler constructors for ingest_completion_event
// and verify_habit_task. Each constructor returns a server.ToolHandlerFunc closure
// wired to real application services following the Clean Architecture DI pattern.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// HandleIngestCompletionEvent returns a ToolHandlerFunc wired to the real IngestService.
// It extracts provider, userID, grindID, and payload from the MCP tool call arguments,
// calls svc.Ingest, and returns the persisted CompletionEvent as JSON text.
//
// Error handling:
//   - ErrHabitTaskNotFound  → mcpgo.NewToolResultError (domain error, not a Go error)
//   - other errors           → mcpgo.NewToolResultError with wrapped message
func HandleIngestCompletionEvent(svc *services.IngestService) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		args := req.GetArguments()
		provider, _ := args["provider"].(string)
		userID, _ := args["userID"].(string)
		grindID, _ := args["grindID"].(string)
		rawPayload, _ := args["payload"].(map[string]interface{})

		event, err := svc.Ingest(provider, userID, grindID, rawPayload)
		if err != nil {
			switch {
			case errors.Is(err, config.ErrHabitTaskNotFound):
				return mcpgo.NewToolResultError("no habit task found for today"), nil
			default:
				return mcpgo.NewToolResultError(fmt.Sprintf("ingestion failed: %v", err)), nil
			}
		}

		out, _ := json.Marshal(event)
		return mcpgo.NewToolResultText(string(out)), nil
	}
}

// HandleVerifyHabitTask returns a ToolHandlerFunc that checks whether a given
// CompletionEvent (by completionEventID) exists among the events associated with
// a HabitTask (by habitTaskID).
//
// Error handling:
//   - habit task not found       → mcpgo.NewToolResultError (domain error)
//   - completion events DB error → return nil, err (infrastructure failure; let MCP server recover)
//   - no matching event          → returns {"verified":false} as text
//   - matching event found       → returns {"verified":true} as text
func HandleVerifyHabitTask(
	habitTaskRepo repositories.HabitTaskRepository,
	completionEventRepo repositories.CompletionEventRepository,
) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		args := req.GetArguments()
		habitTaskID, _ := args["habitTaskID"].(string)
		completionEventID, _ := args["completionEventID"].(string)

		task, err := habitTaskRepo.FindByID(habitTaskID)
		if err != nil {
			return mcpgo.NewToolResultError("habit task not found"), nil
		}

		events, err := completionEventRepo.FindByHabitTaskID(task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch completion events: %w", err)
		}

		for _, e := range events {
			if e.ID == completionEventID {
				out, _ := json.Marshal(map[string]bool{"verified": true})
				return mcpgo.NewToolResultText(string(out)), nil
			}
		}

		return mcpgo.NewToolResultText(`{"verified":false}`), nil
	}
}
