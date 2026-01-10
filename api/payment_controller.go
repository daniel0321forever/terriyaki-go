package api

import (
	"errors"
	"fmt"

	"github.com/daniel0321forever/terriyaki-go/internal/config"
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/daniel0321forever/terriyaki-go/internal/serializer"
	"github.com/daniel0321forever/terriyaki-go/internal/services"
	"github.com/daniel0321forever/terriyaki-go/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
CreateInstantPaymentAPI handles instant payment intent creation via Stripe.

@route	POST	/api/v1/payments/instant
@desc	Create a Stripe payment intent for an instant payment and return the client secret.
@auth	Required

Request Body (application/json):

	{
		"amount": int // Amount in the smallest currency unit (for USD: cents, e.g., 500 = $5.00)
	}

Response [200]:

	{
		"clientSecret": string // The client secret used by the client to complete payment with Stripe.js
	}

Possible Errors:
- 400 Bad Request: Invalid or missing request body.
- 500 Internal Server Error: Failed to create Stripe payment service or payment intent.

Example:
POST /api/v1/payments/instant

	{
		"amount": 1000
	}

RESPONSE:

	{
		"clientSecret": "pi_3JoawMaE5p_lBLABL0SMq_example_secret"
	}
*/
func PaymentIntentAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		Amount int `json:"amount" binding:"required"`
	}
	var body Request

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		c.JSON(400, gin.H{"message": "Invalid request body: must provide integer 'amount'"})
		return
	}

	// create payment service.
	// TODO: get it as a parameter after refactor
	paymentService, err := services.NewStripePaymentService()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, "Internal Server Error")
		return
	}

	// create payment intent
	clientSecret, err := paymentService.CreatePaymentIntent(int64(body.Amount))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, "Internal Server Error")
		return
	}

	// return client secret
	c.JSON(200, gin.H{
		"clientSecret": clientSecret,
	})
}

/*
CreateSaveCardIntentAPI handles the creation of a Stripe setup intent for saving a card.

@route  POST   /api/v1/payments/save-card-intent
@desc   Create a Stripe setup intent and return the client secret for saving a card.
@auth   Required

Response [200]:

	{
	    "clientSecret": string // The client secret used by the client to save a card with Stripe.js
	}

Possible Errors:
- 500 Internal Server Error: Failed to create Stripe payment service or setup intent.

Example:
POST /api/v1/payments/save-card-intent

RESPONSE:

	{
	    "clientSecret": "seti_1N2FaBEXAMPLE_csecret_sample"
	}
*/
func SaveCardIntentAPI(c *gin.Context) {
	// Create payment service (to be refactored to DI)
	paymentService, err := services.NewStripePaymentService()
	if err != nil {
		fmt.Println("Could not create Stripe payment service:", err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// Create save card (setup) intent with Stripe
	clientSecret, err := paymentService.CreateSaveCardIntent()
	if err != nil {
		fmt.Println("Could not create save card intent:", err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// Return the client secret for the setup intent
	c.JSON(200, gin.H{"clientSecret": clientSecret})
}

func SaveCardAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		StripePaymentMethodID string `json:"payment_method_id" binding:"required"`
	}
	var body Request

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		c.JSON(400, gin.H{"message": "Invalid request body: must provide string 'payment_method_id'", "errorCode": config.ERROR_CODE_BAD_REQUEST})
		return
	}

	// get user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized", "errorCode": config.ERROR_CODE_UNAUTHORIZED})
		return
	}

	user, err := models.GetUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(401, gin.H{"message": "User not found", "errorCode": config.ERROR_CODE_USER_NOT_FOUND})
			return
		}
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// create payment service.
	// TODO: get it as a parameter after refactor
	paymentService, err := services.NewStripePaymentService()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// save card
	_, err = paymentService.SaveCard(user, body.StripePaymentMethodID)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// return the success message
	c.JSON(200, gin.H{"message": "Card saved successfully"})
}

/*
ForceInvestigateDuedPenaltyAPI handles the investigation of dued penalties.

@route  POST   /api/v1/charges/dued
@desc   Force investigate dued penalties.
@auth   Required

Response [200]:
*/
func ForceInvestigateDuedPenaltyAPI(c *gin.Context) {
	// authenticate the user
	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized", "errorCode": config.ERROR_CODE_UNAUTHORIZED})
		return
	}

	// create payment service.
	// TODO: get it as a parameter after refactor
	paymentService, err := services.NewStripePaymentService()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// find dued payments
	pendingPayments, err := paymentService.FindDuedPayments()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// charge the pending payments
	for _, pendingPayment := range pendingPayments {
		_, err := paymentService.Charge(pendingPayment.StripePaymentInfo, pendingPayment.PaymentAmount)
		if err != nil {
			fmt.Println(err)
			c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
			return
		}
	}

	// return the success message
	c.JSON(200, gin.H{"message": "Dued penalties charged successfully"})
}

/*
GetAvailablePaymentMethodsAPI handles the retrieval of available payment methods.

@route  GET   /api/v1/payments/available-payment-methods
@desc   Get the available payment methods for the user.
@auth   Required

Response [200]:

	{
		"payment_methods": [
			{
				"payment_method_id": "pm_1234567890",
				"brand": "visa",
				"last4": "4242",
				"exp_month": 12,
				"exp_year": 2025
			}
		],
		"default_payment_method": "pm_1234567890"
	}
*/
func GetAvailablePaymentMethodsAPI(c *gin.Context) {
	// authenticate the user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized", "errorCode": config.ERROR_CODE_UNAUTHORIZED})
		return
	}

	user, err := models.GetUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(401, gin.H{"message": "User not found", "errorCode": config.ERROR_CODE_USER_NOT_FOUND})
			return
		}
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// create payment service.
	// TODO: get it as a parameter after refactor
	paymentService, err := services.NewStripePaymentService()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// get available payment methods
	paymentInfos, defaultPaymentInfo, err := paymentService.GetAvailablePaymentMethods(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// return the available payment methods
	c.JSON(200, gin.H{"payment_methods": serializer.SerializePaymentInfos(paymentInfos), "default_payment_method": defaultPaymentInfo.StripePaymentMethodID})
}

/*
SelectPaymentMethodAPI handles the selection of a payment method.

@route  POST   /api/v1/payments/methods/select-default
@desc   Select a payment method as the default payment method.
@auth   Required

Response [200]:

	{
		"message": "Default payment method updated successfully"
	}
*/
func SelectPaymentMethodAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		StripePaymentMethodID string `json:"payment_method_id" binding:"required"`
	}
	var body Request

	// authenticate the user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized", "errorCode": config.ERROR_CODE_UNAUTHORIZED})
		return
	}

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		c.JSON(400, gin.H{"message": "Invalid request body: must provide string 'stripe_payment_method_id'", "errorCode": config.ERROR_CODE_BAD_REQUEST})
		return
	}

	// update the user's default payment method
	_, err = models.UpdateUser(userID, nil, nil, nil, nil, nil, &body.StripePaymentMethodID)
	if err != nil {
		fmt.Println("Error updating user's default payment method:", err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// return the success message
	c.JSON(200, gin.H{"message": "Default payment method updated successfully"})
}

func TestForceChargingAPI(c *gin.Context) {
	// authenticate the user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized", "errorCode": config.ERROR_CODE_UNAUTHORIZED})
		return
	}

	user, err := models.GetUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(401, gin.H{"message": "User not found", "errorCode": config.ERROR_CODE_USER_NOT_FOUND})
			return
		}
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// create payment service.
	// TODO: get it as a parameter after refactor
	paymentService, err := services.NewStripePaymentService()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// get payment info
	_, defaultPaymentInfo, err := paymentService.GetAvailablePaymentMethods(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}

	// force charge the user
	_, err = paymentService.Charge(*defaultPaymentInfo, 100)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Internal Server Error", "errorCode": config.ERROR_CODE_INTERNAL_SERVER_ERROR})
		return
	}
}
