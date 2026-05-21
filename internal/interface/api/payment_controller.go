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
	userService   *services.UserService
	stripeService services.StripePaymentService
	solanaService services.SolanaPaymentService
}

func NewPaymentController(
	us *services.UserService,
	stripeService services.StripePaymentService,
	solanaService services.SolanaPaymentService,
) *PaymentController {
	return &PaymentController{
		userService:   us,
		stripeService: stripeService,
		solanaService: solanaService,
	}
}

/*
PaymentIntentAPI handles instant payment intent creation via Stripe.

PAYMENT METHOD: Stripe only

@route	POST	/api/v1/payments/stripe/payment-intent
@desc	Create a Stripe payment intent for an instant payment and return the client secret.
@auth	Required

Request Body (application/json):

	{
		"amount_cents": int // Amount in the smallest currency unit (for USD: cents, e.g., 500 = $5.00)
	}

Response [200]:

	{
		"clientSecret": string,
		"idempotent_replay": bool
	}

Example:
POST /api/v1/payments/stripe/payment-intent

	{
		"amount_cents": 1000
	}

RESPONSE:

	{
		"clientSecret": "pi_3JoawMaE5p_lBLABL0SMq_example_secret"
	}
*/
func (ctrl *PaymentController) PaymentIntentAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		AmountCents int `json:"amount_cents" binding:"required"`
	}
	var body Request

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		RespondBadRequest(c, "Invalid request body: must provide integer 'amount_cents'")
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		RespondBadRequest(c, "Missing required Idempotency-Key header")
		return
	}

	if ctrl.stripeService == nil {
		RespondInternalServerError(c, "Stripe payment service is not configured")
		return
	}

	// create payment intent
	result, err := ctrl.stripeService.CreateStripeCollectionIntent(dto.StripeCreateIntentDTO{UserID: "", AmountCents: int64(body.AmountCents)}, idempotencyKey)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// return client secret
	c.JSON(200, gin.H{
		"clientSecret":      result.ClientSecret,
		"idempotent_replay": result.IdempotentReplay,
	})
}

/*
AddPaymentMethodAPI handles provider-agnostic payment method onboarding.

PAYMENT METHOD: Shared onboarding for Stripe cards and Solana wallets

This endpoint intentionally remains unprefixed because it accepts both card-based
and wallet-based onboarding payloads.

@route  POST   /api/v1/payments/methods
@desc   Onboard a payment method for either Stripe cards or Solana wallets.
@auth   Required

Response [200]:

	{
	    "message": string,
	    "payment_method": object
	}

Possible Errors:
- 400 Bad Request: Invalid or missing request body.
- 500 Internal Server Error: Failed to add payment method.

Example:
POST /api/v1/payments/methods

	{
		"method_type": "card",
		"card_payment_method_id": "pm_1234567890"
	}

RESPONSE:

	{
	    "message": "Payment method added successfully",
	    "payment_method": {...}
	}
*/
func (ctrl *PaymentController) AddPaymentMethodAPI(c *gin.Context) {
	// Define struct for expected body format (strict binding)
	type Request struct {
		MethodType          string `json:"method_type" binding:"required"`
		CardPaymentMethodID string `json:"card_payment_method_id"`
		WalletAddress       string `json:"wallet_address"`
		Network             string `json:"network"`
		ProgramID           string `json:"program_id"`
	}
	var body Request

	// get body
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		RespondBadRequest(c, "Invalid request body: must provide 'method_type' and provider-specific fields")
		return
	}

	// get user
	token := c.GetHeader("Authorization")
	userID, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "Unauthorized")
		return
	}

	addMethodDTO, err := dto.NewAddPaymentMethodDTO(
		userID,
		body.MethodType,
		body.CardPaymentMethodID,
		body.WalletAddress,
		body.Network,
		body.ProgramID,
	)
	if err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	var paymentService services.PaymentServiceCore
	switch strings.ToLower(strings.TrimSpace(body.MethodType)) {
	case "card":
		if ctrl.stripeService == nil {
			RespondInternalServerError(c, "Stripe payment service is not configured")
			return
		}
		paymentService = ctrl.stripeService
	case "solana_wallet":
		if ctrl.solanaService == nil {
			RespondInternalServerError(c, "Solana payment service is not configured")
			return
		}
		paymentService = ctrl.solanaService
	default:
		RespondBadRequest(c, "Invalid method_type: supported values are 'card' and 'solana_wallet'")
		return
	}

	// add payment method
	result, err := paymentService.AddPaymentMethod(addMethodDTO)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// return the success message
	c.JSON(200, gin.H{"message": "Payment method added successfully", "payment_method": result.PaymentMethod})
}

/*
ForceInvestigateDuedPenaltyAPI handles the investigation of dued penalties and charges them.

PAYMENT METHOD: Stripe only

@route  POST   /api/v1/payments/stripe/force-charging
@desc   Force investigate dued penalties and charge them via Stripe.
@auth   Required

Response [200]:

	{
		"message": string,
		"reconciled_settlements": int
	}
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

	if ctrl.stripeService == nil {
		RespondInternalServerError(c, "Stripe payment service is not configured")
		return
	}

	// find dued payments
	res, err := ctrl.stripeService.FindDuedPayments()
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// charge the pending payments
	for i, pendingPayment := range res.PendingPayments {
		operationKey := fmt.Sprintf("%s:%d:%s", idempotencyKey, i, pendingPayment.PaymentMethodInfo.ProviderPaymentMethodID)
		chargeReq, err := dto.NewChargeWithIdempotencyDTO(pendingPayment.PaymentMethodInfo, pendingPayment.PaymentAmount, "force_charging", userID)
		if err != nil {
			fmt.Println(err)
			RespondInternalServerError(c, "Internal Server Error")
			return
		}
		_, err = ctrl.stripeService.ChargeWithIdempotency(chargeReq, operationKey)
		if err != nil {
			fmt.Println(err)
			RespondInternalServerError(c, "Internal Server Error")
			return
		}
	}

	reconReq, err := dto.NewReconcileSettlementsDTO(100)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}
	reconciled, err := ctrl.stripeService.ReconcileSettlements(reconReq)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// return the success message
	c.JSON(200, gin.H{
		"message":                "Dued penalties charged successfully",
		"reconciled_settlements": len(reconciled.UpdatedSettlements),
	})
}

/*
GetAvailablePaymentMethodsAPI handles the retrieval of available payment methods.

PAYMENT METHOD: Stripe only

@route  GET   /api/v1/payments/stripe/methods
@desc   Get the available payment methods (cards) for the user via Stripe.
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

	if ctrl.stripeService == nil {
		RespondInternalServerError(c, "Stripe payment service is not configured")
		return
	}

	// get available payment methods
	getReq, err := dto.NewGetAvailablePaymentMethodsDTO(userID)
	if err != nil {
		RespondBadRequest(c, err.Error())
		return
	}
	availablePaymentMethods, err := ctrl.stripeService.GetAvailablePaymentMethods(getReq)
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
SelectPaymentMethodAPI handles the selection of a payment method as default.

PAYMENT METHOD: Stripe only

@route  POST   /api/v1/payments/stripe/methods/select-default
@desc   Select a payment method (card) as the default and charge a verification amount.
@auth   Required

Response [200]:

	{
		"message": "Payment method selected and charged successfully"
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

	if ctrl.stripeService == nil {
		RespondInternalServerError(c, "Stripe payment service is not configured")
		return
	}

	res, err := ctrl.stripeService.ClaimIdempotency("method_selection", idempotencyKey)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}
	if !res.Claimed {
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
	getReq2, err := dto.NewGetAvailablePaymentMethodsDTO(userID)
	if err != nil {
		RespondBadRequest(c, err.Error())
		return
	}
	availablePaymentMethods, err := ctrl.stripeService.GetAvailablePaymentMethods(getReq2)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	// force charge the user
	chargeReq2, err := dto.NewChargeWithIdempotencyDTO(availablePaymentMethods.DefaultPaymentInfo, 100, "method_selection_charge", userID)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}
	_, err = ctrl.stripeService.ChargeWithIdempotency(chargeReq2, idempotencyKey+":"+body.StripePaymentMethodID)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, "Internal Server Error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method selected and charged successfully"})
}

/*
CreateSolanaCollectionIntentAPI creates a Solana unsigned transaction payload for client signing.

PAYMENT METHOD: Solana only (non-custodial)

Generates an unsigned transaction that the client's wallet will sign locally.
Returns the unsigned transaction JSON, pledge PDA, and recent blockhash.

@route  POST   /api/v1/payments/solana/collection-intent
@desc   Create an unsigned Solana transaction for client-side signing.
@auth   Required
*/
func (ctrl *PaymentController) CreateSolanaCollectionIntentAPI(c *gin.Context) {
	if ctrl.solanaService == nil {
		RespondInternalServerError(c, "Solana payment service is not configured")
		return
	}

	type Request = dto.SolanaCreateIntentDTO
	var body Request

	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		RespondBadRequest(c, "Invalid request body: must provide Solana intent fields")
		return
	}

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

	validatedBody, ctorErr := dto.NewSolanaCreateIntentDTO(userID, body.WalletAddress, body.Network, body.ProgramID, body.PledgeID, body.OraclePubkey, body.AmountLamports, body.DeadlineUnix)
	if ctorErr != nil {
		RespondBadRequest(c, ctorErr.Error())
		return
	}

	result, err := ctrl.solanaService.CreateSolanaCollectionIntent(validatedBody, idempotencyKey)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

/*
SubmitSolanaSignedTransactionAPI submits a client-signed Solana transaction for broadcast.

PAYMENT METHOD: Solana only (non-custodial)

Broadcasts the client-signed transaction to the Solana blockchain via RPC,
persists settlement proof (signature, timestamp, network), and updates settlement status.

@route  POST   /api/v1/payments/solana/submit-signed-transaction
@desc   Submit a client-signed Solana transaction for on-chain settlement.
@auth   Required
*/
func (ctrl *PaymentController) SubmitSolanaSignedTransactionAPI(c *gin.Context) {
	if ctrl.solanaService == nil {
		RespondInternalServerError(c, "Solana payment service is not configured")
		return
	}

	type Request = dto.SolanaSubmitSignedTransactionDTO
	var body Request

	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Invalid request body, causing the following error:\n" + err.Error())
		RespondBadRequest(c, "Invalid request body: must provide signed transaction payload")
		return
	}

	token := c.GetHeader("Authorization")
	_, err := utils.VerifyUserAccess(token)
	if err != nil {
		RespondUnauthorized(c, "Unauthorized")
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		RespondBadRequest(c, "Missing required Idempotency-Key header")
		return
	}

	validatedSubmit, ctorErr := dto.NewSolanaSubmitSignedTransactionDTO(body.ProviderReference, body.SignedTransactionBase64, body.Network)
	if ctorErr != nil {
		RespondBadRequest(c, ctorErr.Error())
		return
	}

	result, err := ctrl.solanaService.SubmitSolanaSignedTransaction(validatedSubmit, idempotencyKey)
	if err != nil {
		fmt.Println(err)
		RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}
