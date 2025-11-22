package serializer

import (
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
)

/**
 * Serialize a grind
 * @param grind - the grind to serialize
 * @return the serialized grind
 */
func SerializeGrind(grind *models.Grind) gin.H {
	participantsRecords := []gin.H{}
	for _, participant := range grind.Participants {
		participantsRecords = append(participantsRecords, SerializeUserAsGrindParticipant(&participant, grind))
	}

	// get progress tasks
	progressTasks := SerializeGrindTasks(grind)

	// get today task
	taskToday, _ := models.GetTodayTask(grind.Participants[0].ID, grind.ID)
	taskTodayRecord := gin.H{}
	if taskToday != nil {
		taskTodayRecord = SerializeTask(taskToday)
	}

	return gin.H{
		"id":           grind.ID,
		"duration":     grind.Duration,
		"budget":       grind.Budget,
		"startDate":    grind.StartDate,
		"participants": participantsRecords,
		"taskToday":    taskTodayRecord,
		"progress":     progressTasks,
	}
}

/**
 * Serialize a list of grinds
 * @param grinds - the list of grinds to serialize
 * @return the serialized list of grinds
 */
func SerializeGrinds(grinds []models.Grind) []gin.H {
	grindsRecords := []gin.H{}
	for _, grind := range grinds {
		grindsRecords = append(grindsRecords, SerializeGrind(&grind))
	}
	return grindsRecords
}
