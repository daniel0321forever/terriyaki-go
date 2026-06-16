package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware returns a Gin handler that enforces per-IP rate limiting using Redis.
// It is a stub — implementation follows in the GREEN phase.
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
