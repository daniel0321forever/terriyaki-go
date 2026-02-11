package api

import (
	"net/http"
	"strconv"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
)

type MessageController struct {
	userService    *services.UserService
	messageService *services.MessageService
	grindService   *services.GrindService
}

func NewMessageController(
	us *services.UserService,
	ms *services.MessageService,
	gs *services.GrindService,
) *MessageController {
	return &MessageController{
		userService:    us,
		messageService: ms,
		grindService:   gs,
	}
}

func (ctrl *MessageController) GetMessageAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	_, err = ctrl.userService.GetUser(getUserDTO)
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

	getMessageDTO := dto.GetAllMessagesForReceiverDTO{
		ReceiverID: userID,
		Offset:     offset,
		Limit:      limit,
	}
	messageDTOs, err := ctrl.messageService.GetAllMessagesForReceiver(getMessageDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// Build response with related data
	responseData := []*dto.MessageDTO{}
	for _, messageDTO := range messageDTOs {
		responseData = append(responseData, messageDTO)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Messages fetched successfully",
		"data":    responseData,
	})
}

func (ctrl *MessageController) ReadMessageAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	_, err = ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	messageID := c.Param("id")

	updateMessageDTO := dto.UpdateMessageReadStatusDTO{
		MessageID: messageID,
		Read:      true,
	}

	messageDTO, err := ctrl.messageService.UpdateMessageReadStatus(updateMessageDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message read successfully",
		"data":    messageDTO,
	})
}

func (ctrl *MessageController) CreateInvitationAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	// Get user from user id
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid request body",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	receiverEmail := body["participantEmail"].(string)
	grindID := body["grindID"].(string)

	// Create invitation message
	createMessageDTO := dto.CreateInvitationMessageDTO{
		SenderID:      userID,
		ReceiverEmail: receiverEmail,
		GrindID:       grindID,
	}
	messageDTO, err := ctrl.messageService.CreateInvitationMessage(createMessageDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation created successfully",
		"data":    messageDTO,
	})
}

func (ctrl *MessageController) AcceptInvitationAPI(c *gin.Context) {
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

	// Get user data from user id
	getUserDTO := dto.GetUserDTO{UserID: userID}
	accepterDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// Get message information
	messageID := c.Param("id")
	getMessageDTO := dto.GetMessageDTO{MessageID: messageID}
	messageDTO, err := ctrl.messageService.GetMessageByID(getMessageDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "inviting message not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}
	grindID := messageDTO.InvitationGrind.ID

	// Add participant to grind (still uses string params - service not updated yet)
	addParticipationDTO := dto.AddParticipationDTO{
		GrindID: grindID,
		UserID:  userID,
	}
	err = ctrl.grindService.AddParticipation(addParticipationDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// Get invitor user data from inviting message sender id
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.Sender.ID}
	invitorDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "invitor not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// Update message status
	updateMessageDTO := dto.UpdateMessageInvitationAcceptedStatusDTO{
		MessageID: messageID,
		Accepted:  true,
	}
	_, err = ctrl.messageService.UpdateMessageInvitationAcceptedStatus(updateMessageDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// Create invitation accepted message
	createMessageDTO := dto.CreateInvitationAcceptedMessageDTO{
		AccepterID: accepterDTO.ID,
		InvitorID:  invitorDTO.ID,
		GrindID:    grindID,
	}
	messageDTO, err = ctrl.messageService.CreateInvitationAcceptedMessage(createMessageDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	// Get grind
	getGrindDTO := dto.GetGrindDTO{GrindID: grindID}
	grindDTO, err := ctrl.grindService.GetGrind(getGrindDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation accepted successfully",
		"data": gin.H{
			"message":  messageDTO,
			"sender":   invitorDTO,
			"receiver": accepterDTO,
			"grind":    grindDTO,
		},
	})
}

func (ctrl *MessageController) RejectInvitationAPI(c *gin.Context) {
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

	// Get user data from user id
	getUserDTO := dto.GetUserDTO{UserID: userID}
	rejectorDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "user not found",
			"errorCode": config.ERROR_CODE_USER_NOT_FOUND,
		})
		return
	}

	// Get inviting message id from path parameter
	messageID := c.Param("id")
	getMessageDTO := dto.GetMessageDTO{MessageID: messageID}
	messageDTO, err := ctrl.messageService.GetMessageByID(getMessageDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "inviting message not found",
			"errorCode": config.ERROR_CODE_INVITING_MESSAGE_NOT_FOUND,
		})
		return
	}

	// Get invitor user data from inviting message sender id
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.Sender.ID}
	invitorDTO, _ := ctrl.userService.GetUser(getUserDTO)

	// Update message invitation responded status
	updateMessageDTO := dto.UpdateMessageInvitationAcceptedStatusDTO{
		MessageID: messageID,
		Accepted:  false,
	}
	_, _ = ctrl.messageService.UpdateMessageInvitationAcceptedStatus(updateMessageDTO)

	// Create invitation rejected message
	createMessageDTO := dto.CreateInvitationRejectedMessageDTO{
		RejecterID: rejectorDTO.ID,
		InvitorID:  invitorDTO.ID,
		GrindID:    messageDTO.InvitationGrind.ID,
	}
	messageDTO, _ = ctrl.messageService.CreateInvitationRejectedMessage(createMessageDTO)

	// Get grind
	getGrindDTO := dto.GetGrindDTO{GrindID: messageDTO.InvitationGrind.ID}
	grindDTO, err := ctrl.grindService.GetGrind(getGrindDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation rejected successfully",
		"data": gin.H{
			"message":  messageDTO,
			"sender":   invitorDTO,
			"receiver": rejectorDTO,
			"grind":    grindDTO,
		},
	})
}

/**
 * Get all the messages that the user has sent
 * @param c - the context
 * @return the messages that the user has sent
 */
func (ctrl *MessageController) GetSentMessageAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, _ := strconv.Atoi(offsetStr)
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	messages, err := ctrl.messageService.GetAllMessageFromSender(userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "internal server error",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Messages fetched successfully",
		"data":    messages,
	})
}
