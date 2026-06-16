package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
)

// PartnerGroupController handles partner group CRUD and invite endpoints.
type PartnerGroupController struct {
	groupService *services.PartnerGroupService
}

// NewPartnerGroupController creates a new PartnerGroupController.
func NewPartnerGroupController(groupService *services.PartnerGroupService) *PartnerGroupController {
	return &PartnerGroupController{groupService: groupService}
}

// extractUserID reads the Bearer token from the Authorization header and returns the userID.
// Returns empty string and writes an Unauthorized response if auth fails.
func (ctrl *PartnerGroupController) extractUserID(c *gin.Context) (string, bool) {
	userID, err := utils.VerifyUserAccess(c.GetHeader("Authorization"))
	if err != nil {
		RespondUnauthorized(c, "authentication required")
		return "", false
	}
	return userID, true
}

// CreateGroupAPI handles POST /api/v2/groups.
func (ctrl *PartnerGroupController) CreateGroupAPI(c *gin.Context) {
	userID, ok := ctrl.extractUserID(c)
	if !ok {
		return
	}

	var body dto.CreatePartnerGroupDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondBadRequest(c, "invalid request body")
		return
	}

	group, err := ctrl.groupService.CreateGroup(body.GrindID, userID, body.Name)
	if err != nil {
		RespondInternalServerError(c, "failed to create partner group")
		return
	}

	c.JSON(http.StatusCreated, mappers.BuildPartnerGroupDTO(group))
}

// GetGroupAPI handles GET /api/v2/groups/:id.
func (ctrl *PartnerGroupController) GetGroupAPI(c *gin.Context) {
	_, ok := ctrl.extractUserID(c)
	if !ok {
		return
	}

	group, err := ctrl.groupService.GetGroup(c.Param("id"))
	if err != nil {
		RespondNotFound(c, "partner group not found")
		return
	}

	c.JSON(http.StatusOK, mappers.BuildPartnerGroupDTO(group))
}

// GenerateInviteLinkAPI handles POST /api/v2/groups/:id/invite.
func (ctrl *PartnerGroupController) GenerateInviteLinkAPI(c *gin.Context) {
	userID, ok := ctrl.extractUserID(c)
	if !ok {
		return
	}

	token, err := ctrl.groupService.GenerateInviteToken(c.Param("id"), userID)
	if err != nil {
		if errors.Is(err, config.ErrForbidden) {
			RespondForbidden(c, "only the group owner can generate invite links")
			return
		}
		RespondInternalServerError(c, "failed to generate invite link")
		return
	}

	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	inviteURL := fmt.Sprintf("%s/groups/join?token=%s", frontendBaseURL, token)

	c.JSON(http.StatusOK, dto.InviteLinkDTO{
		Token:     token,
		InviteURL: inviteURL,
	})
}

// JoinGroupAPI handles POST /api/v2/groups/join.
func (ctrl *PartnerGroupController) JoinGroupAPI(c *gin.Context) {
	userID, ok := ctrl.extractUserID(c)
	if !ok {
		return
	}

	var body dto.JoinGroupDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondBadRequest(c, "invalid request body")
		return
	}

	group, err := ctrl.groupService.JoinGroup(body.Token, userID)
	if err != nil {
		// JWT errors include "token is expired" and "token signature is invalid"
		if strings.Contains(err.Error(), "token") {
			RespondBadRequest(c, "invalid or expired invite token")
			return
		}
		RespondInternalServerError(c, "failed to join group")
		return
	}

	c.JSON(http.StatusOK, mappers.BuildPartnerGroupDTO(group))
}
