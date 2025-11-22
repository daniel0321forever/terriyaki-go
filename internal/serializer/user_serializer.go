package serializer

import (
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
)

/**
 * Serialize a user
 * @param user - the user to serialize
 * @return the serialized user, `id`, `username`, `email` and `avatar`
 */
func SerializeUser(user *models.User) gin.H {
	return gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
	}
}

/**
 * Serialize a user as a grind participant
 * @param user - the user to serialize
 * @param grind - the grind to get the missed task and total penalty
 * @return the serialized user as a grind participant
 */
func SerializeUserAsGrindParticipant(user *models.User, grind *models.Grind) gin.H {
	missedTask := 0
	totalPenalty := 0

	tasks := []models.Task{}
	result := database.Db.Where("user_id = ? AND grind_id = ?", user.ID, grind.ID).Order("date ASC").Find(&tasks)
	if result.Error != nil {
		return gin.H{}
	}

	for _, task := range tasks {
		if !task.Completed {
			missedTask++
			totalPenalty += int(grind.Budget / grind.Duration)
		}
		if task.Date.After(time.Now().UTC()) {
			break
		}
	}

	return gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"avatar":       user.Avatar,
		"missedDays":   missedTask,
		"totalPenalty": totalPenalty,
	}
}
