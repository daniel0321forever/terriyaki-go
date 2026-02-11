package api

import (
	"fmt"
	"net/http"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
	}

	taskID := c.Param("id")
	setProblemIfNeeded := c.Query("set-problem") == "true"

	getTaskDTO := dto.GetTaskDTO{
		TaskID:             taskID,
		SetProblemIfNeeded: setProblemIfNeeded,
	}
	taskDTO, err := ctrl.taskService.GetTaskByID(getTaskDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "Task not found",
			"task":      nil,
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userID,
	}
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "Grind not found",
			"task":      nil,
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "Error fetching task",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}
	getGrindDTO := dto.GetOngoingGrindDTO{
		UserID: userID,
	}
	// TODO: might have multiple grinds in "GetOngoingGrindByUserID"
	// perhaps makeing grindId an input to the FinishTodayTaskAPI?
	grindDTO, err := ctrl.grindService.GetOngoingGrindByUserID(getGrindDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "Grind not found",
			"grind":     nil,
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	getTaskDTO := dto.GetTodayTaskDTO{
		UserID:  userID,
		GrindID: grindDTO.ID,
	}
	taskDTO, err := ctrl.taskService.GetTodayTask(getTaskDTO)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "Task not found",
			"task":      nil,
			"errorCode": config.ERROR_CODE_NOT_FOUND,
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

	code, ok := body["code"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid code",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
		return
	}

	codeLanguage, ok := body["language"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":   "invalid code language",
			"errorCode": config.ERROR_CODE_BAD_REQUEST,
		})
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "Error fetching progress records",
			"errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Progress records fetched successfully",
		"progressRecords": progressRecords,
	})
}
