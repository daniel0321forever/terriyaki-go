package services

import (
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
)

func TestPaymentServiceGetAvailablePaymentMethods(t *testing.T) {
	t.Parallel()

	userRepo := new(mocks.MockUserRepository)
	userRepo.On("FindById", "u1").Return(&entities.User{ID: "u1", DefaultPaymentMethodID: "pm_2"}, nil)

	infoRepo := new(mocks.MockStripePaymentInfoRepository)
	infoRepo.On("FindByUserID", "u1").Return([]entities.PaymentMethodInfo{
		{ProviderPaymentMethodID: "pm_1", ProviderCustomerID: "cus_1", Brand: "visa"},
		{ProviderPaymentMethodID: "pm_2", ProviderCustomerID: "cus_2", Brand: "mastercard"},
	}, nil)

	svc := &PaymentService{userRepo: userRepo, paymentMethodInfoRepo: infoRepo}
	getReq, err := dto.NewGetAvailablePaymentMethodsDTO("u1")
	if err != nil {
		t.Fatalf("constructor error: %v", err)
	}
	res, err := svc.GetAvailablePaymentMethods(getReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.PaymentInfos) != 2 {
		t.Fatalf("expected two payment infos")
	}
	if res.DefaultPaymentInfo.ProviderPaymentMethodID != "pm_2" {
		t.Fatalf("expected pm_2 as default, got %q", res.DefaultPaymentInfo.ProviderPaymentMethodID)
	}
	userRepo.AssertExpectations(t)
	infoRepo.AssertExpectations(t)
}

func TestPaymentServiceFindDuedPayments(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	grindRepo.On("FindDuedGrinds").Return([]*entities.Grind{{
		ID:           "g1",
		Participants: []entities.User{{ID: "u1"}},
	}}, nil)

	partRepo := new(mocks.MockParticipationRepository)
	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(&entities.Participation{UserID: "u1", GrindID: "g1", TotalPenalty: 42}, nil)

	infoRepo := new(mocks.MockStripePaymentInfoRepository)
	infoRepo.On("FindByUserID", "u1").Return([]entities.PaymentMethodInfo{{UserID: "u1", ProviderPaymentMethodID: "pm_1", ProviderCustomerID: "cus_1"}}, nil)

	svc := &PaymentService{
		grindRepo:             grindRepo,
		participationRepo:     partRepo,
		paymentMethodInfoRepo: infoRepo,
	}

	result, err := svc.FindDuedPayments()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.PendingPayments) != 1 {
		t.Fatalf("expected one pending payment, got %d", len(result.PendingPayments))
	}
	if result.PendingPayments[0].PaymentAmount != 42 {
		t.Fatalf("expected payment amount 42, got %d", result.PendingPayments[0].PaymentAmount)
	}
	grindRepo.AssertExpectations(t)
	partRepo.AssertExpectations(t)
	infoRepo.AssertExpectations(t)
}
