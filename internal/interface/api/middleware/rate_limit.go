package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware returns a Gin handler factory that enforces per-IP rate limiting
// using a Redis-backed counter.
//
// Strategy: fixed-window counter per client IP + full path.
//   - Key format: "rate:{clientIP}:{fullPath}"
//   - On first request (count == 1): set TTL to window duration.
//   - On count > limit: abort with 429.
//   - On Redis error: fail-open (c.Next()) — Redis unavailability must never block traffic.
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate:%s:%s", ip, c.FullPath())

		count, err := rdb.Incr(c.Request.Context(), key).Result()
		if err != nil {
			// Fail-open: Redis unavailability must not block legitimate traffic.
			c.Next()
			return
		}

		if count == 1 {
			// Set TTL only on the first request to avoid resetting the window on every call.
			// Unconditional Expire is an anti-pattern (T-03-07).
			rdb.Expire(c.Request.Context(), key, window) //nolint:errcheck
		}

		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}
