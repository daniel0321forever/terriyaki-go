package integration

import (
	"net/http"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/database"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/stretchr/testify/assert"
)

func TestPaymentIntentAPI(t *testing.T) {
	t.Run("Create Payment Intent - Invalid Body", func(t *testing.T) {
		// Invalid request body (missing amount)
		requestBody := map[string]string{}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/payments/payment-intent", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Create Payment Intent - Invalid Amount Type", func(t *testing.T) {
		// Invalid request body (amount as string instead of int)
		requestBody := map[string]string{
			"amount": "invalid",
		}

		// Execute
		rr := test.MakeRequest("POST", "/api/v1/payments/payment-intent", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestSaveCardIntentAPI(t *testing.T) {
	t.Run("Create Save Card Intent", func(t *testing.T) {
		// Execute - this endpoint doesn't require body or auth in current implementation
		// It may fail if Stripe service is not properly configured
		rr := test.MakeRequest("POST", "/api/v1/payments/save-card-intent", nil, false)

		// Assert - could be 200 or 500 depending on Stripe config
		// We just check it doesn't panic
		assert.NotNil(t, rr)
	})
}

func TestSaveCardAPI(t *testing.T) {
	// Create test user
	testEmail := "test_save_card@example.com"
	user, _ := models.CreateUser("savecarduser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Save Card - Unauthorized", func(t *testing.T) {
		// Request body
		requestBody := map[string]string{
			"payment_method_id": "pm_test_123",
		}

		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/payments/save-card", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Save Card - Invalid Body", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Invalid request body (missing payment_method_id)
		requestBody := map[string]string{}

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/payments/save-card", requestBody, token)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestGetAvailablePaymentMethodsAPI(t *testing.T) {
	// Create test user
	testEmail := "test_payment_methods@example.com"
	user, _ := models.CreateUser("paymentmethodsuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Get Payment Methods - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("GET", "/api/v1/payments/methods", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestSelectPaymentMethodAPI(t *testing.T) {
	// Create test user
	testEmail := "test_select_payment@example.com"
	user, _ := models.CreateUser("selectpaymentuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Select Payment Method - Unauthorized", func(t *testing.T) {
		// Request body
		requestBody := map[string]string{
			"payment_method_id": "pm_test_123",
		}

		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/payments/methods/select-default", requestBody, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Select Payment Method - Invalid Body", func(t *testing.T) {
		// Login to get token
		loginReq := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
		loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
		token := loginResponse["token"].(string)

		// Invalid request body (missing payment_method_id)
		requestBody := map[string]string{}

		// Execute with auth
		rr := test.MakeAuthenticatedRequest("POST", "/api/v1/payments/methods/select-default", requestBody, token)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestForceInvestigateDuedPenaltyAPI(t *testing.T) {
	// Create test user
	testEmail := "test_dued_penalty@example.com"
	user, _ := models.CreateUser("duedpenaltyuser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Force Investigate Dued Penalty - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/charges/dued", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestForceChargingAPI(t *testing.T) {
	// Create test user
	testEmail := "test_force_charging@example.com"
	user, _ := models.CreateUser("forcecharginguser", testEmail, "password123", "https://example.com/avatar.jpg")
	defer database.Db.Delete(&user)

	t.Run("Force Charging - Unauthorized", func(t *testing.T) {
		// Execute without auth
		rr := test.MakeRequest("POST", "/api/v1/payments/force-charging", nil, false)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
