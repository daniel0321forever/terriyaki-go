package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	userService    *services.UserService
	paymentService services.IPaymentService
}

func NewPaymentController(
	us *services.UserService,
	ps services.IPaymentService,
) *PaymentController {
	return &PaymentController{
		userService:    us,
		paymentService: ps,
	}
}

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
func (ctrl *PaymentController) PaymentIntentAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		Amount int `json:"amount" binding:"required"`
	}
	var body Request

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		RespondBadRequest(c, "Invalid request body: must provide integer 'amount'")
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		RespondBadRequest(c, "Missing required Idempotency-Key header")
		return
	}

	// create payment intent
	clientSecret, replayed, err := ctrl.paymentService.CreatePaymentIntentWithIdempotency(int64(body.Amount), idempotencyKey)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// return client secret
	c.JSON(200, gin.H{
		"clientSecret":      clientSecret,
		"idempotent_replay": replayed,
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
func (ctrl *PaymentController) SaveCardIntentAPI(c *gin.Context) {
	// Create save card (setup) intent with Stripe
	clientSecret, err := ctrl.paymentService.CreateSaveCardIntent()
	if err != nil {
		fmt.Println("Could not create save card intent:", err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// Return the client secret for the setup intent
	c.JSON(200, gin.H{"clientSecret": clientSecret})
}

func (ctrl *PaymentController) SaveCardAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		StripePaymentMethodID string `json:"payment_method_id" binding:"required"`
	}
	var body Request

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		RespondBadRequest(c, "Invalid request body: must provide string 'payment_method_id'")
		return
	}

	// get user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "Unauthorized")
		return
	}

	saveCardDTO := dto.SaveCardDTO{
		UserID:          userID,
		PaymentMethodID: body.StripePaymentMethodID,
	}

	// save card
	_, err = ctrl.paymentService.SaveCard(saveCardDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
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
func (ctrl *PaymentController) ForceInvestigateDuedPenaltyAPI(c *gin.Context) {
	// authenticate the user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "Unauthorized")
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		RespondBadRequest(c, "Missing required Idempotency-Key header")
		return
	}

	// find dued payments
	pendingPayments, err := ctrl.paymentService.FindDuedPayments()
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// charge the pending payments
	for i, pendingPayment := range pendingPayments {
		operationKey := fmt.Sprintf("%s:%d:%s", idempotencyKey, i, pendingPayment.PaymentMethodInfo.ProviderPaymentMethodID)
		_, _, err := ctrl.paymentService.ChargeWithIdempotency(pendingPayment.PaymentMethodInfo, pendingPayment.PaymentAmount, "force_charging", operationKey, userID)
		if err != nil {
			fmt.Println(err)
			RespondInternalServerError(c, "Internal Server Error")
			return
		}
	}

	reconciled, err := ctrl.paymentService.ReconcileSettlements(100)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// return the success message
	c.JSON(200, gin.H{
		"message":                "Dued penalties charged successfully",
		"reconciled_settlements": len(reconciled),
	})
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
func (ctrl *PaymentController) GetAvailablePaymentMethodsAPI(c *gin.Context) {
	// authenticate the user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "Unauthorized")
		return
	}

	// get available payment methods
	availablePaymentMethods, err := ctrl.paymentService.GetAvailablePaymentMethods(dto.GetAvailablePaymentMethodsDTO{UserID: userID})
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_methods":        availablePaymentMethods.PaymentInfos,
		"default_payment_method": availablePaymentMethods.DefaultPaymentInfo.ProviderPaymentMethodID,
	})
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
func (ctrl *PaymentController) SelectPaymentMethodAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		StripePaymentMethodID string `json:"payment_method_id" binding:"required"`
	}
	var body Request

	// authenticate the user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "Unauthorized")
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		RespondBadRequest(c, "Missing required Idempotency-Key header")
		return
	}

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		RespondBadRequest(c, "Invalid request body: must provide string 'stripe_payment_method_id'")
		return
	}

	claimed, err := ctrl.paymentService.ClaimIdempotency("method_selection", idempotencyKey)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}
	if !claimed {
		c.JSON(http.StatusOK, gin.H{"message": "Duplicate method selection ignored by idempotency key"})
		return
	}

	// update the user's default payment method
	_, err = ctrl.userService.UpdateUser(dto.UpdateUserDTO{
		UserID:                 userID,
		DefaultPaymentMethodID: &body.StripePaymentMethodID,
	})
	if err != nil {
		fmt.Println("Error updating user's default payment method:", err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// get payment info
	availablePaymentMethods, err := ctrl.paymentService.GetAvailablePaymentMethods(dto.GetAvailablePaymentMethodsDTO{UserID: userID})
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// force charge the user
	_, _, err = ctrl.paymentService.ChargeWithIdempotency(availablePaymentMethods.DefaultPaymentInfo, 100, "method_selection_charge", idempotencyKey+":"+body.StripePaymentMethodID, userID)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method selected and charged successfully"})
}
