package api

import (
	"crypto/hmac"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
)

// IngestController handles POST /api/v2/ingest/:provider.
type IngestController struct {
	ingestService *services.IngestService
}

// NewIngestController creates a new IngestController.
func NewIngestController(ingestService *services.IngestService) *IngestController {
	return &IngestController{ingestService: ingestService}
}

// HandleIngest processes an external habit completion signal.
// Supports two auth modes:
//   - Bearer <jwt>: Chrome-extension user token — userID extracted from JWT.
//   - ApiKey <key>: B2B webhook token — constant-time compared against INGEST_API_KEY;
//     userID must be supplied in the request body.
func (ctrl *IngestController) HandleIngest(c *gin.Context) {
	// Enforce 1 MB body limit before parsing (mitigates T-02-09 DoS via large JSONB payload).
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)

	authHeader := c.GetHeader("Authorization")
	var userID string

	switch {
	case strings.HasPrefix(authHeader, "Bearer "):
		uid, err := utils.VerifyUserAccess(authHeader)
		if err != nil {
			RespondUnauthorized(c, "invalid or missing credentials")
			return
		}
		userID = uid

	case strings.HasPrefix(authHeader, "ApiKey "):
		expectedKey := os.Getenv("INGEST_API_KEY")
		if expectedKey == "" {
			// If the env var is unset, ApiKey auth is unavailable (T-02-04).
			RespondError(c, http.StatusServiceUnavailable, config.ERROR_CODE_INTERNAL_SERVER_ERROR, "API key authentication is not configured")
			return
		}
		providedKey := strings.TrimPrefix(authHeader, "ApiKey ")
		// Constant-time comparison to prevent timing attacks (T-02-04).
		if !hmac.Equal([]byte(providedKey), []byte(expectedKey)) {
			RespondUnauthorized(c, "invalid or missing credentials")
			return
		}
		// For B2B webhooks the caller supplies userID in the body; defer extraction below.

	default:
		RespondUnauthorized(c, "invalid or missing credentials")
		return
	}

	var rawBody map[string]interface{}
	if err := c.ShouldBindJSON(&rawBody); err != nil {
		RespondBadRequest(c, "invalid request body")
		return
	}

	// For ApiKey mode, userID comes from the request body.
	if strings.HasPrefix(authHeader, "ApiKey ") {
		uid, ok := rawBody["userID"].(string)
		if !ok || uid == "" {
			RespondBadRequest(c, "userID is required for API key authentication")
			return
		}
		userID = uid
	}

	grindID, ok := rawBody["grindID"].(string)
	if !ok || grindID == "" {
		RespondBadRequest(c, "grindID is required")
		return
	}

	provider := c.Param("provider")
	event, err := ctrl.ingestService.Ingest(provider, userID, grindID, rawBody)
	if err != nil {
		switch {
		case errors.Is(err, config.ErrHabitTaskNotFound):
			RespondNotFound(c, "no habit task found for today")
		case strings.Contains(err.Error(), "unsupported provider"):
			RespondBadRequest(c, err.Error())
		default:
			RespondInternalServerError(c, "ingestion failed")
		}
		return
	}

	c.JSON(http.StatusCreated, mappers.BuildCompletionEventDTO(event))
}
