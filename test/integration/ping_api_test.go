package integration

import (
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

func TestPingAPI(t *testing.T) {
	t.Run("Successful Ping", func(t *testing.T) {
		// Execute
		rr := test.MakeRequest("GET", "/api/v1/ping", nil, false)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		response := test.ParseResponseMap(t, rr.Body.Bytes())
		expected := map[string]interface{}{
			"message": "pong",
		}
		test.AssertMapValues(t, response, expected, "")
	})
}
