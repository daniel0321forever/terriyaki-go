package api

import (
	"net/http"
	"strconv"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
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
		RespondUnauthorized(c, "unauthorized")
		return
	}

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	_, err = ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		RespondUnauthorized(c, "user not found")
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
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
		RespondInternalServerError(c, "internal server error")
		return
	}

	// Build response with related data
	responseData := []*dto.MessageDTO{}
	responseData = append(responseData, messageDTOs...)

	c.JSON(http.StatusOK, responseData)
}

func (ctrl *MessageController) ReadMessageAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	_, err = ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		RespondNotFound(c, "user not found")
		return
	}

	messageID := c.Param("id")

	updateMessageDTO := dto.UpdateMessageReadStatusDTO{
		MessageID: messageID,
		Read:      true,
	}

	messageDTO, err := ctrl.messageService.UpdateMessageReadStatus(updateMessageDTO)
	if err != nil {
		RespondInternalServerError(c, "internal server error")
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
		RespondUnauthorized(c, "unauthorized")
		return
	}

	// Get user from user id
	body := map[string]any{}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondBadRequest(c, "invalid request body")
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
		RespondInternalServerError(c, "internal server error")
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
		RespondUnauthorized(c, "unauthorized")
		return
	}

	// Get user data from user id
	getUserDTO := dto.GetUserDTO{UserID: userID}
	accepterDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		RespondNotFound(c, "user not found")
		return
	}

	// Get message information
	messageID := c.Param("id")
	getMessageDTO := dto.GetMessageDTO{MessageID: messageID}
	messageDTO, err := ctrl.messageService.GetMessageByID(getMessageDTO)
	if err != nil {
		RespondNotFound(c, "inviting message not found")
		return
	}
	grindID := messageDTO.InvitationGrind.ID

	// Get invitor user data from inviting message sender id
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.Sender.ID}
	invitorDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		RespondNotFound(c, "invitor not found")
		return
	}

	// Build DTOs for the atomic AcceptInvitation call
	addParticipationDTO := dto.AddParticipationDTO{
		GrindID: grindID,
		UserID:  userID,
	}
	updateMessageDTO := dto.UpdateMessageInvitationAcceptedStatusDTO{
		MessageID: messageID,
		Accepted:  true,
	}
	createAcceptedMsgDTO := dto.CreateInvitationAcceptedMessageDTO{
		AccepterID: accepterDTO.ID,
		InvitorID:  invitorDTO.ID,
		GrindID:    grindID,
	}

	// Single transactional call that creates participation + habit tasks +
	// updates invitation message + creates accepted notification
	if err := ctrl.grindService.AcceptInvitation(
		addParticipationDTO,
		updateMessageDTO,
		createAcceptedMsgDTO,
		ctrl.grindService.MessageRepo(),
	); err != nil {
		RespondInternalServerError(c, "internal server error")
		return
	}

	// Get grind for response
	getGrindDTO := dto.GetGrindDTO{GrindID: grindID}
	grindDTO, err := ctrl.grindService.GetGrind(getGrindDTO)
	if err != nil {
		RespondInternalServerError(c, "internal server error")
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
		RespondUnauthorized(c, "unauthorized")
		return
	}

	// Get user data from user id
	getUserDTO := dto.GetUserDTO{UserID: userID}
	rejectorDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		RespondNotFound(c, "user not found")
		return
	}

	// Get inviting message id from path parameter
	messageID := c.Param("id")
	getMessageDTO := dto.GetMessageDTO{MessageID: messageID}
	messageDTO, err := ctrl.messageService.GetMessageByID(getMessageDTO)
	if err != nil {
		RespondNotFound(c, "inviting message not found")
		return
	}

	// Get invitor user data from inviting message sender id
	getUserDTO = dto.GetUserDTO{UserID: messageDTO.Sender.ID}
	invitorDTO, _ := ctrl.userService.GetUser(getUserDTO)

	// Build DTOs for the atomic RejectInvitationTx call
	updateMessageDTO := dto.UpdateMessageInvitationAcceptedStatusDTO{
		MessageID: messageID,
		Accepted:  false,
	}
	createRejectedMsgDTO := dto.CreateInvitationRejectedMessageDTO{
		RejecterID: rejectorDTO.ID,
		InvitorID:  invitorDTO.ID,
		GrindID:    messageDTO.InvitationGrind.ID,
	}

	if err := ctrl.messageService.RejectInvitationTx(updateMessageDTO, createRejectedMsgDTO); err != nil {
		RespondInternalServerError(c, "internal server error")
		return
	}

	// Get grind for response
	getGrindDTO := dto.GetGrindDTO{GrindID: messageDTO.InvitationGrind.ID}
	grindDTO, err := ctrl.grindService.GetGrind(getGrindDTO)
	if err != nil {
		RespondInternalServerError(c, "internal server error")
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
		RespondUnauthorized(c, "unauthorized")
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, _ := strconv.Atoi(offsetStr)
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	messages, err := ctrl.messageService.GetAllMessageFromSender(userID, offset, limit)
	if err != nil {
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, messages)
}
