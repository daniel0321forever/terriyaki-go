package postgres

import (
	"errors"
	"strings"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

type PaymentIdempotencySchema struct {
	gorm.Model
	Operation string `json:"operation" gorm:"not null;index:idx_payment_idempotency_op_key,unique"`
	Key       string `json:"key" gorm:"not null;index:idx_payment_idempotency_op_key,unique"`
	Response  string `json:"response" gorm:""`
}

func (PaymentIdempotencySchema) TableName() string { return "payment_idempotency_keys" }

type PaymentSettlementSchema struct {
	gorm.Model
	UserID            string `json:"user_id" gorm:"not null"`
	Operation         string `json:"operation" gorm:"not null"`
	IdempotencyKey    string `json:"idempotency_key" gorm:"not null;index:idx_payment_settlement_op_key,unique"`
	Provider          string `json:"provider" gorm:"not null"`
	PaymentMethodID   string `json:"payment_method_id" gorm:"not null"`
	Status            string `json:"status" gorm:"not null"`
	Amount            int64  `json:"amount" gorm:"not null"`
	Currency          string `json:"currency" gorm:"not null"`
	RetryCount        int    `json:"retry_count" gorm:"not null;default:0"`
	LastError         string `json:"last_error" gorm:""`
	ProviderReference string `json:"provider_reference" gorm:""`
	Network           string `json:"network" gorm:""`
	TxHash            string `json:"tx_hash" gorm:""`
	ContractAddress   string `json:"contract_address" gorm:""`
	SettlementProof   string `json:"settlement_proof" gorm:""`
	FinalizedAtUnix   int64  `json:"finalized_at_unix" gorm:""`
}

func (PaymentSettlementSchema) TableName() string { return "payment_settlements" }

type GormPaymentIdempotencyRepository struct {
	db *gorm.DB
}

func NewGormPaymentIdempotencyRepository(db *gorm.DB) *GormPaymentIdempotencyRepository {
	return &GormPaymentIdempotencyRepository{db: db}
}

func (r *GormPaymentIdempotencyRepository) Claim(operation string, key string) (bool, error) {
	model := PaymentIdempotencySchema{Operation: operation, Key: key}
	err := r.db.Create(&model).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return false, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
		return false, nil
	}
	return false, err
}

func (r *GormPaymentIdempotencyRepository) GetResponse(operation string, key string) (string, error) {
	var model PaymentIdempotencySchema
	err := r.db.Where("operation = ? AND key = ?", operation, key).First(&model).Error
	if err != nil {
		return "", err
	}
	return model.Response, nil
}

func (r *GormPaymentIdempotencyRepository) SetResponse(operation string, key string, response string) error {
	return r.db.Model(&PaymentIdempotencySchema{}).
		Where("operation = ? AND key = ?", operation, key).
		Update("response", response).Error
}

type GormPaymentSettlementRepository struct {
	db *gorm.DB
}

func NewGormPaymentSettlementRepository(db *gorm.DB) *GormPaymentSettlementRepository {
	return &GormPaymentSettlementRepository{db: db}
}

func (r *GormPaymentSettlementRepository) Create(settlement *entities.PaymentSettlement) (*entities.PaymentSettlement, error) {
	model := PaymentSettlementSchema{
		UserID:            settlement.UserID,
		Operation:         settlement.Operation,
		IdempotencyKey:    settlement.IdempotencyKey,
		Provider:          string(settlement.Provider),
		PaymentMethodID:   settlement.PaymentMethodID,
		Status:            string(settlement.Status),
		Amount:            settlement.Amount,
		Currency:          settlement.Currency,
		RetryCount:        settlement.RetryCount,
		LastError:         settlement.LastError,
		ProviderReference: settlement.Reference.ProviderReference,
		Network:           settlement.Reference.Network,
		TxHash:            settlement.Reference.TxHash,
		ContractAddress:   settlement.Reference.ContractAddress,
		SettlementProof:   settlement.Reference.SettlementProof,
		FinalizedAtUnix:   settlement.Reference.FinalizedAtUnix,
	}
	if err := r.db.Create(&model).Error; err != nil {
		return nil, err
	}
	settlement.ID = model.ID
	settlement.CreatedAt = model.CreatedAt
	settlement.UpdatedAt = model.UpdatedAt
	return settlement, nil
}

func (r *GormPaymentSettlementRepository) Update(settlement *entities.PaymentSettlement) (*entities.PaymentSettlement, error) {
	updates := map[string]any{
		"status":             string(settlement.Status),
		"retry_count":        settlement.RetryCount,
		"last_error":         settlement.LastError,
		"provider_reference": settlement.Reference.ProviderReference,
		"network":            settlement.Reference.Network,
		"tx_hash":            settlement.Reference.TxHash,
		"contract_address":   settlement.Reference.ContractAddress,
		"settlement_proof":   settlement.Reference.SettlementProof,
		"finalized_at_unix":  settlement.Reference.FinalizedAtUnix,
	}
	result := r.db.Model(&PaymentSettlementSchema{}).Where("id = ?", settlement.ID).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return settlement, nil
}

func (r *GormPaymentSettlementRepository) FindByOperationAndKey(operation string, idempotencyKey string) (*entities.PaymentSettlement, error) {
	var model PaymentSettlementSchema
	if err := r.db.Where("operation = ? AND idempotency_key = ?", operation, idempotencyKey).First(&model).Error; err != nil {
		return nil, err
	}
	return mapSettlementSchemaToEntity(model), nil
}

func (r *GormPaymentSettlementRepository) FindByStatuses(statuses []entities.SettlementStatus, limit int) ([]entities.PaymentSettlement, error) {
	if len(statuses) == 0 {
		return []entities.PaymentSettlement{}, nil
	}
	values := make([]string, 0, len(statuses))
	for _, status := range statuses {
		values = append(values, string(status))
	}

	query := r.db.Where("status IN ?", values).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []PaymentSettlementSchema
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]entities.PaymentSettlement, 0, len(models))
	for _, model := range models {
		result = append(result, *mapSettlementSchemaToEntity(model))
	}
	return result, nil
}

func mapSettlementSchemaToEntity(model PaymentSettlementSchema) *entities.PaymentSettlement {
	return &entities.PaymentSettlement{
		ID:              model.ID,
		UserID:          model.UserID,
		Operation:       model.Operation,
		IdempotencyKey:  model.IdempotencyKey,
		Provider:        entities.PaymentProvider(model.Provider),
		PaymentMethodID: model.PaymentMethodID,
		Status:          entities.SettlementStatus(model.Status),
		Amount:          model.Amount,
		Currency:        model.Currency,
		RetryCount:      model.RetryCount,
		LastError:       model.LastError,
		Reference: entities.SettlementReference{
			ProviderReference: model.ProviderReference,
			Network:           model.Network,
			TxHash:            model.TxHash,
			ContractAddress:   model.ContractAddress,
			SettlementProof:   model.SettlementProof,
			FinalizedAtUnix:   model.FinalizedAtUnix,
		},
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
