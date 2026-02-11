package api

import (
	"net/http"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	if err := c.ShouldBindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	updateUserDTO := dto.UpdateUserDTO{
		UserID:   userID,
		Username: Request.Username,
		Avatar:   Request.Avatar,
	}

	userDTO, err := ctrl.userService.UpdateUser(updateUserDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   err.Error(),
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Username updated successfully",
		"user":    userDTO,
	})
}
