package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/serializer"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid request body",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	duration := int(body["duration"].(float64))
	budget := int(body["budget"].(float64))
	participants := body["participants"].([]interface{})
	startDate, _ := time.Parse(time.RFC3339, body["startDate"].(string))

	grind, err := models.CreateGrind(duration, budget, participants, startDate)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grind created successfully",
		"grind":   serializer.SerializeGrind(grind),
	})
}

func GetUserCurrentGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "Unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	grind, err := models.GetOngoingGrindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message":   "No current grind found",
				"grind":     nil,
				"errorCode": config.ERROR_CODE_NOT_FOUND,
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grind fetched successfully",
		"grind":   serializer.SerializeGrind(grind),
	})
}

func GetAllUserGrindsAPI(c *gin.Context) {
	// TODO: get all grinds for a user
}

func UpdateGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	grindID := c.Param("id")

	duration, _ := strconv.Atoi(c.PostForm("duration"))
	budget, _ := strconv.Atoi(c.PostForm("budget"))

	grind, err := models.UpdateGrind(grindID, map[string]any{
		"duration": duration,
		"budget":   budget,
	})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grind updated successfully", "grind": grind})
}

func DeleteGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	grindID := c.Param("id")
	err = models.DeleteGrind(grindID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grind deleted successfully"})
}

func DeleteAllGrindsAPI(c *gin.Context) {
	// TODO: remove it after testing
	err := models.DeleteAllGrinds()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All grinds deleted successfully"})
}
