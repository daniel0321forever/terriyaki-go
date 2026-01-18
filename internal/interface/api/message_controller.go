package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
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
	responseData := []gin.H{}
	for _, messageDTO := range messageDTOs {

		if len(messageDTO.SenderID) == 0 ||
			len(messageDTO.ReceiverID) == 0 ||
			len(messageDTO.InvitationGrindID) == 0 {
			panic(errors.New("Empty SenderID/ReceiverID/GrindID"))
		}

		// Get sender
		getUserDTO := dto.GetUserDTO{UserID: messageDTO.SenderID}
		senderDTO, _ := ctrl.userService.GetUser(getUserDTO)

		// Get receiver
		getUserDTO = dto.GetUserDTO{UserID: messageDTO.ReceiverID}
		receiverDTO, _ := ctrl.userService.GetUser(getUserDTO)

		// Get grind
		getGrindDTO := dto.GetGrindDTO{GrindID: messageDTO.InvitationGrindID}
		grindDTO, _ := ctrl.grindService.GetGrind(getGrindDTO)

		responseData = append(responseData, gin.H{
			"message":  messageDTOs,
			"sender":   senderDTO,
			"receiver": receiverDTO,
			"grind":    grindDTO,
		})
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

	// Get related data
	if len(messageDTO.SenderID) == 0 ||
		len(messageDTO.ReceiverID) == 0 ||
		len(messageDTO.InvitationGrindID) == 0 {
		panic(errors.New("Empty SenderID/ReceiverID/GrindID"))
	}
	var senderDTO *dto.UserDTO
	var receiverDTO *dto.UserDTO
	var grindDTO *dto.GroupGrindDTO

	// Get sender
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.SenderID}
	senderDTO, _ = ctrl.userService.GetUser(getUserDTO)

	// Get receiver
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.ReceiverID}
	receiverDTO, _ = ctrl.userService.GetUser(getUserDTO)

	// Get grind
	getGrindDTO := dto.GetGrindDTO{GrindID: messageDTO.InvitationGrindID}
	grindDTO, _ = ctrl.grindService.GetGrind(getGrindDTO)

	c.JSON(http.StatusOK, gin.H{
		"message": "Message read successfully",
		"data": gin.H{
			"message":  messageDTO,
			"sender":   senderDTO,
			"receiver": receiverDTO,
			"grind":    grindDTO,
		},
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
	getUserDTO := dto.GetUserDTO{UserID: userID}
	senderDTO, err := ctrl.userService.GetUser(getUserDTO)
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

	grindID, ok := body["grindID"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid grindID",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	participantEmail, ok := body["participantEmail"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid participantEmail",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	getReceiverDTO := dto.GetUserByEmailDTO{Email: participantEmail}
	receiverDTO, err := ctrl.userService.GetUserByEmail(getReceiverDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "participant email not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	getGrindDTO := dto.GetGrindDTO{GrindID: grindID}
	grindDTO, err := ctrl.grindService.GetGrind(getGrindDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "grind not found",
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	// Create invitation message
	createMessageDTO := dto.CreateInvitationMessageDTO{
		SenderID:   senderDTO.ID,
		ReceiverID: receiverDTO.ID,
		GrindID:    grindDTO.ID,
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
		"data": gin.H{
			"message":  messageDTO,
			"sender":   senderDTO,
			"receiver": receiverDTO,
			"grind":    grindDTO,
		},
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
	grindID := messageDTO.InvitationGrindID

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
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.SenderID}
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
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.SenderID}
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
		GrindID:    messageDTO.InvitationGrindID,
	}
	messageDTO, _ = ctrl.messageService.CreateInvitationRejectedMessage(createMessageDTO)

	// Get grind
	getGrindDTO := dto.GetGrindDTO{GrindID: messageDTO.InvitationGrindID}
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
