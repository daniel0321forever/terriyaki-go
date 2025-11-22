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

	if task.Date.Before(time.Now().AddDate(0, 0, 1).UTC()) {
		if task.Completed {
			status = "completed"
		} else {
			status = "missed"
		}
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
func SerializeGrindTasks(grind *models.Grind) []gin.H {
	tasks := []models.Task{}
	result := database.Db.Where("grind_id = ?", grind.ID).Order("date ASC").Find(&tasks)
	if result.Error != nil {
		return []gin.H{}
	}

	tasksRecords := []gin.H{}
	for _, task := range tasks {
		tasksRecords = append(tasksRecords, SerializeTaskAsProgressRecord(&task))
	}
	return tasksRecords
}
