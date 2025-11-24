package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	user, err := models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "user not found",
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

	grind, err := models.CreateGrind(duration, budget, []any{user.Email}, startDate)
	if err != nil {
		if strings.Contains(err.Error(), config.ERROR_CODE_USER_NOT_FOUND) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message":   "participant not found",
				"errorCode": config.ERROR_CODE_USER_NOT_FOUND,
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

	// send invitation messages to all participants
	for _, participant := range participants {
		// skip if the participant is the same as the user
		if participant.(string) == user.Email {
			continue
		}

		participantUser, err := models.GetUserByEmail(participant.(string))
		if err != nil {
			models.DeleteGrind(grind.ID)

			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message":   "Participant " + participant.(string) + " not found",
				"errorCode": config.ERROR_CODE_USER_NOT_FOUND,
			})
			return
		}

		_, err = models.CreateInvitationMessage(user, participantUser, grind.ID)
		if err != nil {
			fmt.Println(err)
			continue // skip if the invitation message is not created
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grind created successfully",
		"grind":   serializer.SerializeGrind(user, grind, false),
	})
}

func GetGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
	}

	user, err := models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
	}

	grindID := c.Param("id")
	fmt.Println("grind id", grindID)
	grind, err := models.GetGrind(grindID)
	if err != nil || grind == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "grind not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
	}

	fmt.Println("aa grind", grind)

	c.JSON(http.StatusOK, gin.H{
		"message": "Grind fetched successfully",
		"grind":   serializer.SerializeGrind(user, grind, false),
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

	user, err := models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
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
		"grind":   serializer.SerializeGrind(user, grind, false),
	})
}

func GetAllUserGrindsAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	grinds, err := models.GetAllUserGrinds(userID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	user, err := models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grinds fetched successfully",
		"grinds":  serializer.SerializeGrinds(user, grinds, true),
	})
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

func QuitGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	grindID := c.Param("id")
	grind, err := models.GetGrind(grindID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "grind not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	participateRecord, err := models.GetParticipateRecordByUserIDAndGrindID(userID, grindID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "participate record not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	participateRecord, err = models.UpdateParticipateRecord(participateRecord.ID, map[string]any{
		"totalPenalty": grind.Budget,
		"quitted":      true,
		"quittedAt":    time.Now(),
	})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Grind quitted successfully",
		"participateRecord": serializer.SerializeParticipateRecord(participateRecord),
	})
}
