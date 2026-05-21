package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
)

// GrindController holds the dependencies for all Grind-related endpoints
type GrindController struct {
	grindService   *services.GrindService
	userService    *services.UserService
	messageService *services.MessageService
}

// NewGrindController creates a new instance with injected services
func NewGrindController(
	gs *services.GrindService,
	us *services.UserService,
	ms *services.MessageService,
) *GrindController {
	return &GrindController{
		grindService:   gs,
		userService:    us,
		messageService: ms,
	}
}

func (ctrl *GrindController) CreateGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	// Get the user that is creating the grind
	getUserDTO := dto.GetUserDTO{
		UserID: userID,
	}
	userDTO, err := ctrl.userService.GetUser(getUserDTO)
	if err != nil {
		RespondUnauthorized(c, "user not found")
		return
	}

	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println(err)
		RespondBadRequest(c, "invalid request body")
		return
	}

	var duration int
	switch v := body["duration"].(type) {
	case string:
		parsed, parseErr := strconv.Atoi(v)
		if parseErr != nil {
			RespondBadRequest(c, "invalid duration")
			return
		}
		duration = parsed
	case float64:
		duration = int(v)
	default:
		RespondBadRequest(c, "invalid duration")
		return
	}

	budgetValue, ok := body["budget"].(float64)
	if !ok {
		RespondBadRequest(c, "invalid budget")
		return
	}
	budget := int(budgetValue)

	startDateText, ok := body["startDate"].(string)
	if !ok {
		RespondBadRequest(c, "invalid startDate")
		return
	}
	startDate, err := time.Parse(time.RFC3339, startDateText)
	if err != nil {
		RespondBadRequest(c, "invalid startDate")
		return
	}
	createGrindDTO := dto.CreateGrindDTO{
		CreatorID: userID,
		Duration:  duration,
		Budget:    budget,
		StartDate: startDate,
	}
	grindDTO, err := ctrl.grindService.CreateGroupGrind(createGrindDTO)

	// Convert participant emails to slice of strings
	participants, _ := body["participants"].([]interface{})
	participantEmails := make([]string, 0, len(participants))
	for _, p := range participants {
		if email, ok := p.(string); ok {
			participantEmails = append(participantEmails, email)
		}
	}

	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	// send invitation messages to all participants
	// FIXME: not supposed to have entities exposed in controller level
	var participantUsers []*dto.UserDTO
	for _, participantEmail := range participantEmails {
		// skip if the participant is the same as the user
		if participantEmail == userDTO.Email {
			continue
		}

		getUserDTO := dto.GetUserByEmailDTO{
			Email: participantEmail,
		}
		participantUser, err := ctrl.userService.GetUserByEmail(getUserDTO)
		participantUsers = append(participantUsers, participantUser)
		if err != nil {
			deleteGrindDTO := dto.DeleteGrindDTO{
				GrindID: userDTO.ID,
			}
			ctrl.grindService.DeleteGrind(deleteGrindDTO)

			RespondNotFound(c, "Participant "+participantEmail+" not found")
			return
		}

		createMessageDTO := dto.CreateInvitationMessageDTO{
			SenderID:      userID,
			ReceiverEmail: participantUser.Email,
			GrindID:       grindDTO.ID,
		}
		_, err = ctrl.messageService.CreateInvitationMessage(createMessageDTO)
		if err != nil {
			fmt.Println(err)
			continue // skip if the invitation message is not created
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grind created successfully",
		"grind":   grindDTO,
	})
}

func (ctrl *GrindController) GetGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	grindID := c.Param("id")
	getGrindDTO := dto.GetGrindDTO{
		GrindID: grindID,
		UserID:  userID,
	}
	grindDTO, err := ctrl.grindService.GetGrind(getGrindDTO)
	if err != nil || grindDTO == nil {
		RespondNotFound(c, "grind not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grind": grindDTO,
	})
}

func (ctrl *GrindController) GetUserCurrentGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "Unauthorized")
		return
	}

	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userID,
	}
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)
	if err != nil {
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grind": grindDTO,
	})
}

func (ctrl *GrindController) GetAllUserGrindsAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	getGrindsDTO := dto.GetAllUserGrindsDTO{
		UserID: userID,
	}

	grindDTOs, err := ctrl.grindService.GetAllUserGrinds(getGrindsDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	ids := make([]string, 0, len(grindDTOs))
	for id := range grindDTOs {
		ids = append(ids, id)
	}
	// Sort grind IDs in ascending lexicographic order.
	sort.Strings(ids)

	grinds := make([]*dto.GroupGrindDTO, 0, len(ids))
	for _, id := range ids {
		grinds = append(grinds, grindDTOs[id])
	}

	c.JSON(http.StatusOK, grinds)
}

func (ctrl *GrindController) UpdateGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	grindID := c.Param("id")

	duration, _ := strconv.Atoi(c.PostForm("duration"))
	budget, _ := strconv.Atoi(c.PostForm("budget"))
	updateGrindDTO := dto.UpdateGrindDTO{
		GrindID:  grindID,
		UserID:   userID,
		Duration: duration,
		Budget:   budget,
	}

	grind, err := ctrl.grindService.UpdateGrind(updateGrindDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grind updated successfully", "grind": grind})
}

func (ctrl *GrindController) DeleteGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	grindID := c.Param("id")
	deleteGrindDTO := dto.DeleteGrindDTO{
		GrindID: grindID,
	}
	err = ctrl.grindService.DeleteGrind(deleteGrindDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grind deleted successfully"})
}

func (ctrl *GrindController) DeleteAllGrindsAPI(c *gin.Context) {
	// TODO: remove it after testing
	err := ctrl.grindService.DeleteAllGrinds()
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All grinds deleted successfully"})
}

// TODO: Fix this to fit new structure
func (ctrl *GrindController) QuitGrindAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	grindID := c.Param("id")
	quitGrindDTO := dto.QuitGrindDTO{
		UserID:  userID,
		GrindID: grindID,
	}

	participationDTO, err := ctrl.grindService.QuitGrind(quitGrindDTO)
	if err != nil {
		RespondInternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Grind quitted successfully",
		"participation": participationDTO,
	})
}
