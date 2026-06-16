// Package mcp defines MCP (Model Context Protocol) tool schemas for Habitat.
// Tool definitions live alongside their handlers so the interface layer is self-contained.
package mcp

import (
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// IngestCompletionEventTool defines the schema for ingesting a habit completion signal
// from an external provider (e.g., leetcode, duolingo, custom).
var IngestCompletionEventTool = mcpgo.NewTool("ingest_completion_event",
	mcpgo.WithDescription("Ingest a habit completion signal from an external provider"),
	mcpgo.WithString("provider",
		mcpgo.Description("The external habit provider identifier"),
		mcpgo.Enum("leetcode", "duolingo", "custom"),
		mcpgo.Required(),
	),
	mcpgo.WithString("userID",
		mcpgo.Description("The ID of the user completing the habit"),
		mcpgo.Required(),
	),
	mcpgo.WithString("grindID",
		mcpgo.Description("The ID of the grind (habit group) the completion belongs to"),
		mcpgo.Required(),
	),
	mcpgo.WithObject("payload",
		mcpgo.Description("Provider-specific completion payload data"),
		mcpgo.Required(),
	),
)

// VerifyHabitTaskTool defines the schema for verifying a completion event against
// a HabitTask constraint.
var VerifyHabitTaskTool = mcpgo.NewTool("verify_habit_task",
	mcpgo.WithDescription("Verify a completion event against a HabitTask constraint"),
	mcpgo.WithString("habitTaskID",
		mcpgo.Description("The ID of the HabitTask to verify against"),
		mcpgo.Required(),
	),
	mcpgo.WithString("completionEventID",
		mcpgo.Description("The ID of the CompletionEvent to verify"),
		mcpgo.Required(),
	),
)
