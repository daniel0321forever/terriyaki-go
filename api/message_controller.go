package api

import (
	"net/http"
	"strconv"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/serializer"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetMessageAPI(c *gin.Context) {
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	offsetStr := c.Query("offset")
	offset, _ := strconv.Atoi(offsetStr)
	limitStr := c.Query("limit")
	limit, _ := strconv.Atoi(limitStr)

	messages, err := models.GetAllMessageForReceiver(userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	serializedMessages := []gin.H{}
	for _, message := range messages {
		serializedMessages = append(serializedMessages, serializer.SerializeMessage(&message))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Messages fetched successfully",
		"data":    serializedMessages,
	})
}

func ReadMessageAPI(c *gin.Context) {
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

	messageID := c.Param("id")

	message, err := models.UpdateMessageReadStatus(messageID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message read successfully",
		"data":    serializer.SerializeMessage(message),
	})
}

func CreateInvitationAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	// get user from user id
	user, err := models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid request body",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	grindID := body["grindID"].(string)
	participantEmail := body["participantEmail"].(string)
	participantUser, err := models.GetUserByEmail(participantEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "participant email not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	grind, err := models.GetGrind(grindID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "grind not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	message, err := models.CreateInvitationMessage(user, participantUser, grind.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation created successfully",
		"data":    serializer.SerializeMessage(message),
	})
}

func AcceptInvitationAPI(c *gin.Context) {
	// verify user access
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	// get user data from user id
	user, err := models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// get message information
	invitingMessageID := c.Param("id")
	invitingMessage, err := models.GetMessageByID(invitingMessageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "inviting message not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// add participant to grind
	err = models.AddParticipantToGrind(invitingMessage.InvitationGrindID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// get invitor user data from inviting message sender id
	invitorUser, err := models.GetUser(invitingMessage.SenderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "invitor not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// update message status
	_, err = models.UpdateMessageInvitationAcceptedStatus(invitingMessageID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// create invitation accepted message
	message, err := models.CreateInvitationAcceptedMessage(user, invitorUser, invitingMessage.InvitationGrindID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// return the invitation accepted message
	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation accepted successfully",
		"data":    serializer.SerializeMessage(message),
	})
}

func RejectInvitationAPI(c *gin.Context) {
	// authorize user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	// get user data from user id
	user, err := models.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_USER_NOT_FOUND,
		})
		return
	}

	// get inviting message id from path parameter
	invitingMessageID := c.Param("id")
	invitingMessage, err := models.GetMessageByID(invitingMessageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "inviting message not found",
			"errorCode": config.ERROR_CODE_INVITING_MESSAGE_NOT_FOUND,
		})
		return
	}

	// get invitor user data from inviting message sender id
	invitorUser, _ := models.GetUser(invitingMessage.SenderID)

	// update message invitation responded status
	_, _ = models.UpdateMessageInvitationAcceptedStatus(invitingMessageID, false)

	// create invitation rejected message
	message, _ := models.CreateInvitationRejectedMessage(user, invitorUser, invitingMessage.InvitationGrindID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation rejected successfully",
		"data":    serializer.SerializeMessage(message),
	})
}

/**
 * Get all the messages that the user has sent
 * @param c - the context
 * @return the messages that the user has sent
 */
func GetSentMessageAPI(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	offset, _ := strconv.Atoi(offsetStr)
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

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

	messages, err := models.GetAllMessageFromSender(userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	serializedMessages := []gin.H{}
	for _, message := range messages {
		serializedMessages = append(serializedMessages, serializer.SerializeMessage(&message))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Messages fetched successfully",
		"data":    serializedMessages,
	})
}
