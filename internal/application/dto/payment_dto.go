package dto

import "github.com/daniel0321forever/terriyaki-go/internal/domain/entities"

type SaveCardDTO struct {
	UserID                string `json:"user_id" binding:"required"`
	StripePaymentMethodID string `json:"payment_method_id" binding:"required"`
}

type GetAvailablePaymentMethodsDTO struct {
	UserID string `json:"user_id" binding:"required"`
}

/*
Output DTOs
*/
type AvailablePaymentMethodsDTO struct {
	PaymentInfos       []entities.StripePaymentInfo `json:"payment_infos"`
	DefaultPaymentInfo entities.StripePaymentInfo   `json:"default_payment_info"`
}
