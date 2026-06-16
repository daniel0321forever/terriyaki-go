package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthController struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewHealthController(db *gorm.DB, rdb *redis.Client) *HealthController {
	return &HealthController{db: db, rdb: rdb}
}

// HealthAPI performs live PostgreSQL and Redis pings, returning structured JSON status.
// Returns 200 when both dependencies are reachable; 503 when either is unreachable.
// Only "ok" or "error" string values are returned — raw errors are never exposed (T-03-01).
func (ctrl *HealthController) HealthAPI(c *gin.Context) {
	pgStatus := "ok"
	if err := ctrl.db.Raw("SELECT 1").Error; err != nil {
		pgStatus = "error"
	}

	redisStatus := "ok"
	if ctrl.rdb == nil {
		// nil rdb means Redis is not initialized — report error (not panic)
		redisStatus = "error"
	} else if err := ctrl.rdb.Ping(c.Request.Context()).Err(); err != nil {
		redisStatus = "error"
	}

	if pgStatus == "ok" && redisStatus == "ok" {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"postgres": pgStatus,
			"redis":    redisStatus,
		})
		return
	}

	c.JSON(http.StatusServiceUnavailable, gin.H{
		"status":   "degraded",
		"postgres": pgStatus,
		"redis":    redisStatus,
	})
}
