package serializer

import (
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
)

/**
 * Serialize a participant
 * @param participant - the participant to serialize
 * @return the serialized participant
 */
/**
 * Serialize a user as a grind participant
 * @param user - the user to serialize
 * @param grind - the grind to get the missed task and total penalty
 * @return the serialized user as a grind participant
 */
func SerializeParticipateRecord(participateRecord *models.ParticipateRecord) gin.H {
	return gin.H{
		"id":           participateRecord.UserID,
		"grindID":      participateRecord.GrindID,
		"missedDays":   participateRecord.MissedDays,
		"totalPenalty": participateRecord.TotalPenalty,
		"quitted":      participateRecord.Quitted,
		"quittedAt":    participateRecord.QuittedAt,
	}
}
