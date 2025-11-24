package serializer

import (
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
)

/**
 * Serialize a message
 * @param message - the message to serialize
 * @return the serialized message
 */
func SerializeMessage(message *models.Message) gin.H {
	sender, err := models.GetUser(message.SenderID)
	if err != nil {
		return gin.H{}
	}
	receiver, err := models.GetUser(message.ReceiverID)
	if err != nil {
		return gin.H{}
	}

	grind, err := models.GetGrind(message.InvitationGrindID)
	if err != nil {
		return gin.H{}
	}

	return gin.H{
		"id":                 message.ID,
		"sender":             SerializeUser(sender),
		"receiver":           SerializeUser(receiver),
		"grind":              SerializeGrind(sender, grind, true),
		"content":            message.Content,
		"type":               message.Type,
		"read":               message.Read,
		"invitationAccepted": message.InvitationAccepted,
		"invitationRejected": message.InvitationRejected,
		"createdAt":          message.CreatedAt,
	}
}
