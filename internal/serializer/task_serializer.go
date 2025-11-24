package serializer

import (
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
)

/**
 * Serialize a task
 * @param task - the task to serialize
 * @return the serialized task
 */
func SerializeTask(task *models.Task) gin.H {
	return gin.H{
		"id":           task.ID,
		"userID":       task.UserID,
		"grindID":      task.GrindID,
		"type":         task.TaskType,
		"date":         task.Date,
		"finishedTime": task.FinishedTime,
		"completed":    task.Completed,
		"code":         task.Code,
		"language":     task.CodeLanguage,
		"title":        task.ProblemTitle,
		"description":  task.ProblemDescription,
		"url":          task.ProblemURL,
		"difficulty":   task.ProblemDifficulty,
		"topicTags":    task.ProblemTopicTags,
	}
}

/**
 * Serialize a task as a progress record, only include date, finished time and status
 * @param task - the task to serialize
 * @return the serialized task as a progress record, `date`, `finishedTime` and `status`
 * @status:
 * - `pending`: the task is not completed
 * - `completed`: the task is completed
 * - `missed`: the task is missed
 */
func SerializeTaskAsProgressRecord(task *models.Task) gin.H {
	status := "pending"
	if task.Completed {
		status = "completed"
	}

	startOfTaskDay := task.Date.UTC().Truncate(24 * time.Hour).Add(time.Hour * 1)
	startOfToday := time.Now().UTC().Truncate(24 * time.Hour)
	if startOfTaskDay.Before(startOfToday) {
		status = "missed"
	}

	return gin.H{
		"id":           task.ID,
		"date":         task.Date,
		"finishedTime": task.FinishedTime,
		"status":       status,
	}
}

/*
*
  - Serialize a list of tasks for a grind
  - @param tasks - the list of tasks to serialize
  - @return the serialized list of tasks
  - `pending`: the task is not completed
  - - `completed`: the task is completed
  - - `missed`: the task is missed
*/
func SerializeGrindTasksForUser(user *models.User, grind *models.Grind) []gin.H {
	tasks := []models.Task{}
	result := database.Db.Where("user_id = ? AND grind_id = ?", user.ID, grind.ID).Order("date ASC").Find(&tasks)
	if result.Error != nil {
		return []gin.H{}
	}

	tasksRecords := []gin.H{}
	for _, task := range tasks {
		tasksRecords = append(tasksRecords, SerializeTaskAsProgressRecord(&task))
	}
	return tasksRecords
}
