package api

import (
	"net/http"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the standardized error response structure following OpenAPI spec
type ErrorResponse struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

// RespondError sends a standardized error response with the given status code and error details
func RespondError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, ErrorResponse{
		Message:   message,
		ErrorCode: code,
	})
}

// RespondUnauthorized sends a 401 Unauthorized error response
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, config.ERROR_CODE_UNAUTHORIZED, message)
}

// RespondBadRequest sends a 400 Bad Request error response
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, config.ERROR_CODE_BAD_REQUEST, message)
}

// RespondNotFound sends a 404 Not Found error response
func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, config.ERROR_CODE_NOT_FOUND, message)
}

// RespondConflict sends a 409 Conflict error response (e.g., duplicate entry)
func RespondConflict(c *gin.Context, message string) {
	RespondError(c, http.StatusConflict, config.ERROR_CODE_DUPLICATE_ENTRY, message)
}

// RespondInternalServerError sends a 500 Internal Server Error response
func RespondInternalServerError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, config.ERROR_CODE_INTERNAL_SERVER_ERROR, message)
}

// RespondForbidden sends a 403 Forbidden error response
func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, config.ERROR_CODE_FORBIDDEN, message)
}

// RespondUnprocessableEntity sends a 422 Unprocessable Entity error response
func RespondUnprocessableEntity(c *gin.Context, message string) {
	RespondError(c, http.StatusUnprocessableEntity, config.ERROR_CODE_UNPROCESSABLE_ENTITY, message)
}
