package serializer

import (
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
	participateRecord, err := models.GetParticipateRecordByUserIDAndGrindID(user.ID, grind.ID)
	if err != nil {
		return gin.H{}
	}

	var serializedParticipateRecord gin.H = SerializeParticipateRecord(participateRecord)
	serializedParticipateRecord["username"] = user.Username

	// Get today's task completion status for this participant
	todayTask, err := models.GetTodayTask(user.ID, grind.ID)
	if err != nil {
		// No task found or error - mark as not completed
		serializedParticipateRecord["todayTaskCompleted"] = false
	} else {
		serializedParticipateRecord["todayTaskCompleted"] = todayTask.Completed
	}

	return serializedParticipateRecord
}
