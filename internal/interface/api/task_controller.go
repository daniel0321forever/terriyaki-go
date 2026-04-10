package api

import (
	"fmt"
	"net/http"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskController struct {
	taskService  *services.TaskService
	grindService *services.GrindService
}

func NewTaskController(
	ts *services.TaskService,
	gs *services.GrindService,
) *TaskController {
	return &TaskController{
		taskService:  ts,
		grindService: gs,
	}
}

func (ctrl *TaskController) GetTaskAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	taskID := c.Param("id")
	setProblemIfNeeded := c.Query("set-problem") == "true"

	getTaskDTO := dto.GetTaskDTO{
		TaskID:             taskID,
		SetProblemIfNeeded: setProblemIfNeeded,
	}
	taskDTO, err := ctrl.taskService.GetTaskByID(getTaskDTO)
	if err != nil {
		RespondNotFound(c, "Task not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task found",
		"task":    taskDTO,
	})
}

func (ctrl *TaskController) GetTodayTaskAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userID,
	}
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)
	if err != nil {
		RespondNotFound(c, "Grind not found")
		return
	}
	getTaskDTO := dto.GetTodayTaskDTO{
		GrindID: grindDTO.ID,
	}
	taskDTO, err := ctrl.taskService.GetTodayTask(getTaskDTO)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"message": "No task found for today",
				"task":    nil,
			})
			return
		}
		RespondInternalServerError(c, "Error fetching task")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task found",
		"task":    taskDTO,
	})
}

func (ctrl *TaskController) FinishTodayTaskAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}
	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userID,
	}
	// TODO: might have multiple grinds in "GetOngoingGrindByUserID"
	// perhaps makeing grindId an input to the FinishTodayTaskAPI?
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)
	if err != nil {
		RespondNotFound(c, "Grind not found")
		return
	}

	getTaskDTO := dto.GetTodayTaskDTO{
		UserID:  userID,
		GrindID: grindDTO.ID,
	}
	taskDTO, err := ctrl.taskService.GetTodayTask(getTaskDTO)
	if err != nil {
		RespondNotFound(c, "Task not found")
		return
	}

	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println(err)
		RespondBadRequest(c, "invalid request body")
		return
	}

	code, ok := body["code"].(string)
	if !ok {
		RespondBadRequest(c, "invalid code")
		return
	}

	codeLanguage, ok := body["language"].(string)
	if !ok {
		RespondBadRequest(c, "invalid code language")
		return
	}

	updateTaskDTO := dto.FinishTaskDTO{
		TaskID:       taskDTO.ID,
		Code:         code,
		CodeLanguage: codeLanguage,
	}
	ctrl.taskService.FinishTask(updateTaskDTO)

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"task":    taskDTO,
	})
}

func (ctrl *TaskController) GetProgressRecordsAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "unauthorized")
		return
	}

	grindID := c.Param("id")
	participantID := c.Query("participantId")
	if participantID == "" {
		participantID = userID
	}

	getTaskProgressListDTO := dto.GetTaskProgressListDTO{
		ParticipationID: participantID,
		GrindID:         grindID,
	}

	progressRecords, err := ctrl.taskService.GetTaskProgressList(getTaskProgressListDTO)
	if err != nil {
		RespondInternalServerError(c, "Error fetching progress records")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Progress records fetched successfully",
		"progressRecords": progressRecords,
	})
}
