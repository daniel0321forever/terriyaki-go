package api

import (
	"net/http"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/serializer"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
)

func UpdateProfileAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	_, err = models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	username := body["username"]
	avatar := body["avatar"]

	user, err := models.UpdateUser(userID, &username, nil, &avatar, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Username updated successfully",
		"user":    serializer.SerializeUser(user),
	})
}
