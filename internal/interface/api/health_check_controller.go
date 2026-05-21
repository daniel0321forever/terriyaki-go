package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (ctrl *HealthController) PingAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
