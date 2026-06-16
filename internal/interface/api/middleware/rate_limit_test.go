package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/interface/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimitUnderLimit: mock INCR returns 1; Expire expected once; request passes (200).
func TestRateLimitUnderLimit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	key := "rate:192.0.2.1:/test"
	mock.ExpectIncr(key).SetVal(1)
	mock.ExpectExpire(key, time.Minute).SetVal(true)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	rl := middleware.RateLimitMiddleware(db, 10, time.Minute)
	r.GET("/test", rl, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRateLimitExceedsLimit: mock INCR returns limit+1; request returns 429.
func TestRateLimitExceedsLimit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	key := "rate:192.0.2.1:/test"
	mock.ExpectIncr(key).SetVal(11) // limit=10, so 11 > 10

	gin.SetMode(gin.TestMode)
	r := gin.New()
	rl := middleware.RateLimitMiddleware(db, 10, time.Minute)
	r.GET("/test", rl, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRateLimitWindowNotReset: mock INCR returns 5 (not 1); Expire NOT called.
func TestRateLimitWindowNotReset(t *testing.T) {
	db, mock := redismock.NewClientMock()
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	key := "rate:192.0.2.1:/test"
	mock.ExpectIncr(key).SetVal(5) // count=5, not 1, so Expire should NOT be called

	gin.SetMode(gin.TestMode)
	r := gin.New()
	rl := middleware.RateLimitMiddleware(db, 10, time.Minute)
	r.GET("/test", rl, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRateLimitRedisError: mock INCR returns error; request passes through (fail-open, 200).
func TestRateLimitRedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	key := "rate:192.0.2.1:/test"
	mock.ExpectIncr(key).SetErr(assert.AnError)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	rl := middleware.RateLimitMiddleware(db, 10, time.Minute)
	r.GET("/test", rl, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
