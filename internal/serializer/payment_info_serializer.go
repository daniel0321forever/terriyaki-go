package serializer

import (
	"github.com/daniel0321forever/terriyaki-go/internal/models"
	"github.com/gin-gonic/gin"
)

func SerializePaymentInfo(paymentInfo *models.StripePaymentInfo) gin.H {
	return gin.H{
		"payment_method_id": paymentInfo.StripePaymentMethodID,
		"brand":             paymentInfo.Brand,
		"last4":             paymentInfo.Last4,
		"exp_month":         paymentInfo.ExpMonth,
		"exp_year":          paymentInfo.ExpYear,
	}
}

func SerializePaymentInfos(paymentInfos []models.StripePaymentInfo) []gin.H {
	var serializedPaymentInfos []gin.H
	for _, paymentInfo := range paymentInfos {
		serializedPaymentInfos = append(serializedPaymentInfos, SerializePaymentInfo(&paymentInfo))
	}
	return serializedPaymentInfos
}
