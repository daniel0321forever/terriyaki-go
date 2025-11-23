package serializer

import (
	"fmt"

	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/**
 * Serialize a grind
 * @param grind - the grind to serialize
 * @return the serialized grind
 */
func SerializeGrind(user *models.User, grind *models.Grind, simple bool) gin.H {
	participantsData := []gin.H{}
	for _, participant := range grind.Participants {
		participantsData = append(participantsData, SerializeUserAsGrindParticipant(&participant, grind))
	}

	// quitted or not
	quitted := false
	participateRecord, err := models.GetParticipateRecordByUserIDAndGrindID(user.ID, grind.ID)
	if err != nil {
		fmt.Println(err)
	}
	quitted = participateRecord.Quitted

	if simple {
		return gin.H{
			"id":           grind.ID,
			"duration":     grind.Duration,
			"budget":       grind.Budget,
			"startDate":    grind.StartDate,
			"participants": participantsData,
			"quitted":      quitted,
		}
	}

	// get progress tasks
	progressTasks := SerializeGrindTasks(grind)

	// get today task
	taskToday, err := models.GetTodayTask(user.ID, grind.ID)
	var taskTodayRecord gin.H

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			taskTodayRecord = gin.H{}
		} else {
			panic(err)
		}
	} else {
		taskTodayRecord = SerializeTask(taskToday)
	}

	return gin.H{
		"id":           grind.ID,
		"duration":     grind.Duration,
		"budget":       grind.Budget,
		"startDate":    grind.StartDate,
		"participants": participantsData,
		"taskToday":    taskTodayRecord,
		"progress":     progressTasks,
		"quitted":      quitted,
	}
}

/**
 * Serialize a list of grinds
 * @param grinds - the list of grinds to serialize
 * @return the serialized list of grinds
 */
func SerializeGrinds(user *models.User, grinds []models.Grind, simple bool) []gin.H {
	grindsRecords := []gin.H{}
	for _, grind := range grinds {
		grindsRecords = append(grindsRecords, SerializeGrind(user, &grind, simple))
	}
	return grindsRecords
}
