package api

import (
	"net/http"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	userService *services.UserService
}

func NewProfileController(us *services.UserService) *ProfileController {
	return &ProfileController{
		userService: us,
	}
}

func (ctrl *ProfileController) UpdateProfileAPI(c *gin.Context) {
	Request := struct {
		Username *string `json:"username"`
		Avatar   *string `json:"avatar"`
	}{}

	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	if err := c.ShouldBindJSON(&Request); err != nil {
		RespondBadRequest(c, "invalid request body")
		return
	}

	updateUserDTO := dto.UpdateUserDTO{
		UserID:   userID,
		Username: Request.Username,
		Avatar:   Request.Avatar,
	}

	userDTO, err := ctrl.userService.UpdateUser(updateUserDTO)
	if err != nil {
		RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Username updated successfully",
		"user":    userDTO,
	})
}
