package services

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

type fakePaymentAdapter struct {
	chargeResponse string
	chargeErr      error
}

func (f *fakePaymentAdapter) CreatePaymentIntent(amount int64) (string, error) {
	return fmt.Sprintf("pi_%d", amount), nil
}

func (f *fakePaymentAdapter) CreateSaveCardIntent() (string, error) {
	return "seti_test", nil
}

func (f *fakePaymentAdapter) CreateCustomer(name string, email string) (string, error) {
	return "cus_test", nil
}

func (f *fakePaymentAdapter) DescribePaymentMethod(paymentMethodID string) (*entities.PaymentMethodInfo, error) {
	return entities.NewPaymentMethodInfo(entities.PaymentProviderStripe, "", "", paymentMethodID, "visa", "4242", 1, 2030), nil
}

func (f *fakePaymentAdapter) AttachPaymentMethodToCustomer(paymentMethodID string, customerID string) error {
	return nil
}

func (f *fakePaymentAdapter) Charge(customerID string, paymentMethodID string, amount int64) (string, error) {
	if f.chargeErr != nil {
		return "", f.chargeErr
	}
	if f.chargeResponse != "" {
		return f.chargeResponse, nil
	}
	return "pi_charge_test", nil
}

func (f *fakePaymentAdapter) PayBack(destinationAccountID string, amount int64) error {
	return nil
}

type inMemoryIdempotencyRepo struct {
	mu   sync.Mutex
	data map[string]string
}

func newInMemoryIdempotencyRepo() *inMemoryIdempotencyRepo {
	return &inMemoryIdempotencyRepo{data: map[string]string{}}
}

func (r *inMemoryIdempotencyRepo) Claim(operation string, key string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := operation + "::" + key
	if _, exists := r.data[k]; exists {
		return false, nil
	}
	r.data[k] = ""
	return true, nil
}

func (r *inMemoryIdempotencyRepo) GetResponse(operation string, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := operation + "::" + key
	v, exists := r.data[k]
	if !exists {
		return "", errors.New("not found")
	}
	return v, nil
}

func (r *inMemoryIdempotencyRepo) SetResponse(operation string, key string, response string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := operation + "::" + key
	if _, exists := r.data[k]; !exists {
		return errors.New("not found")
	}
	r.data[k] = response
	return nil
}

type inMemorySettlementRepo struct {
	mu   sync.Mutex
	next uint
	data map[string]*entities.PaymentSettlement
}

func newInMemorySettlementRepo() *inMemorySettlementRepo {
	return &inMemorySettlementRepo{next: 1, data: map[string]*entities.PaymentSettlement{}}
}

func settlementKey(operation string, idempotencyKey string) string {
	return operation + "::" + idempotencyKey
}

func (r *inMemorySettlementRepo) Create(settlement *entities.PaymentSettlement) (*entities.PaymentSettlement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := settlementKey(settlement.Operation, settlement.IdempotencyKey)
	if _, exists := r.data[k]; exists {
		return nil, errors.New("duplicate settlement")
	}
	copySettlement := *settlement
	copySettlement.ID = r.next
	copySettlement.CreatedAt = time.Now().Add(-11 * time.Minute)
	copySettlement.UpdatedAt = copySettlement.CreatedAt
	r.next++
	r.data[k] = &copySettlement
	result := copySettlement
	return &result, nil
}

func (r *inMemorySettlementRepo) Update(settlement *entities.PaymentSettlement) (*entities.PaymentSettlement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := settlementKey(settlement.Operation, settlement.IdempotencyKey)
	if _, exists := r.data[k]; !exists {
		return nil, errors.New("settlement not found")
	}
	copySettlement := *settlement
	copySettlement.UpdatedAt = time.Now()
	r.data[k] = &copySettlement
	result := copySettlement
	return &result, nil
}

func (r *inMemorySettlementRepo) FindByOperationAndKey(operation string, idempotencyKey string) (*entities.PaymentSettlement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, exists := r.data[settlementKey(operation, idempotencyKey)]
	if !exists {
		return nil, errors.New("not found")
	}
	copySettlement := *v
	return &copySettlement, nil
}

func (r *inMemorySettlementRepo) FindByStatuses(statuses []entities.SettlementStatus, limit int) ([]entities.PaymentSettlement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	allowed := map[entities.SettlementStatus]struct{}{}
	for _, s := range statuses {
		allowed[s] = struct{}{}
	}
	result := make([]entities.PaymentSettlement, 0)
	for _, settlement := range r.data {
		if _, ok := allowed[settlement.Status]; ok {
			result = append(result, *settlement)
		}
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

func TestPaymentIntentIdempotency(t *testing.T) {
	t.Parallel()

	adapter := &fakePaymentAdapter{}
	idempotencyRepo := newInMemoryIdempotencyRepo()
	settlementRepo := newInMemorySettlementRepo()

	svc := &PaymentService{adapter: adapter, idempotencyRepo: idempotencyRepo, settlementRepo: settlementRepo}

	first, replayed, err := svc.CreatePaymentIntentWithIdempotency(500, "key-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if replayed {
		t.Fatalf("first call should not be replay")
	}

	second, replayed, err := svc.CreatePaymentIntentWithIdempotency(500, "key-1")
	if err != nil {
		t.Fatalf("expected no error on replay, got %v", err)
	}
	if !replayed {
		t.Fatalf("second call should be replay")
	}
	if first != second {
		t.Fatalf("expected same response on replay, got %q and %q", first, second)
	}
}

func TestChargeSettlementLifecycleAndReconciliation(t *testing.T) {
	t.Parallel()

	adapter := &fakePaymentAdapter{chargeErr: errors.New("provider timeout")}
	idempotencyRepo := newInMemoryIdempotencyRepo()
	settlementRepo := newInMemorySettlementRepo()

	svc := &PaymentService{adapter: adapter, idempotencyRepo: idempotencyRepo, settlementRepo: settlementRepo}
	paymentInfo := entities.PaymentMethodInfo{
		Provider:                entities.PaymentProviderStripe,
		ProviderCustomerID:      "cus_1",
		ProviderPaymentMethodID: "pm_1",
	}

	_, replayed, err := svc.ChargeWithIdempotency(paymentInfo, 100, "force_charging", "charge-key-1", "u1")
	if err == nil {
		t.Fatalf("expected charge error")
	}
	if replayed {
		t.Fatalf("first failed charge should not be replay")
	}

	settlement, findErr := settlementRepo.FindByOperationAndKey("force_charging", "charge-key-1")
	if findErr != nil {
		t.Fatalf("expected settlement, got error: %v", findErr)
	}
	if settlement.Status != entities.SettlementStatusFailed {
		t.Fatalf("expected failed status, got %q", settlement.Status)
	}

	reconciled, err := svc.ReconcileSettlements(10)
	if err != nil {
		t.Fatalf("expected no reconcile error, got %v", err)
	}
	if len(reconciled) == 0 {
		t.Fatalf("expected reconciled settlements")
	}

	settlement, _ = settlementRepo.FindByOperationAndKey("force_charging", "charge-key-1")
	if settlement.Status != entities.SettlementStatusPending {
		t.Fatalf("expected pending after retry scheduling, got %q", settlement.Status)
	}

	adapter.chargeErr = nil
	adapter.chargeResponse = "pi_capture_123"
	ref, replayed, err := svc.ChargeWithIdempotency(paymentInfo, 100, "force_charging", "charge-key-2", "u1")
	if err != nil {
		t.Fatalf("expected successful charge, got %v", err)
	}
	if replayed {
		t.Fatalf("new idempotency key should not replay")
	}
	if ref != "pi_capture_123" {
		t.Fatalf("expected provider reference pi_capture_123, got %q", ref)
	}

	captured, _ := settlementRepo.FindByOperationAndKey("force_charging", "charge-key-2")
	if captured.Status != entities.SettlementStatusCaptured {
		t.Fatalf("expected captured status, got %q", captured.Status)
	}

	refReplay, replayed, err := svc.ChargeWithIdempotency(paymentInfo, 100, "force_charging", "charge-key-2", "u1")
	if err != nil {
		t.Fatalf("expected replay success, got %v", err)
	}
	if !replayed {
		t.Fatalf("expected replay on duplicate key")
	}
	if refReplay != ref {
		t.Fatalf("expected replay reference %q, got %q", ref, refReplay)
	}
}
