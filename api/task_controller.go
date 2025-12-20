package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/serializer"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetTaskAPI(c *gin.Context) {
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

	task, err := models.GetTaskByID(taskID, setProblemIfNeeded)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "Task not found",
			"task":      nil,
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task found", "task": serializer.SerializeTask(task)})
}

func GetTodayTaskAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	grind, err := models.GetOngoingGrindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "Grind not found",
			"task":      nil,
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	task, err := models.GetTodayTask(userID, grind.ID)
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

	c.JSON(http.StatusOK, gin.H{"message": "Task found", "task": serializer.SerializeTask(task)})
}

func FinishTodayTaskAPI(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":   "unauthorized",
			"errorCode": config.ERROR_CODE_UNAUTHORIZED,
		})
		return
	}

	grind, err := models.GetOngoingGrindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "Grind not found",
			"grind":     nil,
			"errorCode": config.ERROR_CODE_NOT_FOUND,
		})
		return
	}

	task, err := models.GetTodayTask(userID, grind.ID)
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

	// TODO: mark task as completed by checking recent accepted problems via leetcode account
	// submissions, err := leetcodeapi.GetUserRecentAcSubmissions(testUsername, pageSize) and task.ProblemTitle
	task.Completed = true
	task.FinishedTime = time.Now()
	task.Code = &code
	task.CodeLanguage = &codeLanguage

	database.Db.Save(task)

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "task": task})
}
