package services

import (
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
	"github.com/stretchr/testify/mock"
)

func TestStripePaymentServiceSelectPaymentMethod(t *testing.T) {
	t.Parallel()

	userRepo := new(mocks.MockUserRepository)
	userRepo.On("Update", mock.MatchedBy(func(u *entities.User) bool {
		return u.ID == "u1" && u.DefaultPaymentMethodID == "pm_new"
	})).Return(nil)

	svc := &StripePaymentService{userRepo: userRepo}

	user := &entities.User{ID: "u1", DefaultPaymentMethodID: "pm_old"}
	err := svc.SelectPaymentMethod(user, "pm_new")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.DefaultPaymentMethodID != "pm_new" {
		t.Fatalf("expected default payment method updated")
	}
	userRepo.AssertExpectations(t)
}

func TestStripePaymentServiceGetAvailablePaymentMethods(t *testing.T) {
	t.Parallel()

	userRepo := new(mocks.MockUserRepository)
	userRepo.On("FindById", "u1").Return(&entities.User{ID: "u1", DefaultPaymentMethodID: "pm_2"}, nil)

	infoRepo := new(mocks.MockStripePaymentInfoRepository)
	infoRepo.On("FindByUserID", "u1").Return([]entities.StripePaymentInfo{
		{StripePaymentMethodID: "pm_1", Brand: "visa"},
		{StripePaymentMethodID: "pm_2", Brand: "mastercard"},
	}, nil)

	svc := &StripePaymentService{userRepo: userRepo, stripePaymentInfoRepo: infoRepo}
	res, err := svc.GetAvailablePaymentMethods(dto.GetAvailablePaymentMethodsDTO{UserID: "u1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.PaymentInfos) != 2 {
		t.Fatalf("expected two payment infos")
	}
	if res.DefaultPaymentInfo.StripePaymentMethodID != "pm_2" {
		t.Fatalf("expected pm_2 as default, got %q", res.DefaultPaymentInfo.StripePaymentMethodID)
	}
	userRepo.AssertExpectations(t)
	infoRepo.AssertExpectations(t)
}

func TestStripePaymentServiceFindDuedPayments(t *testing.T) {
	t.Parallel()

	grindRepo := new(mocks.MockGrindRepository)
	grindRepo.On("FindDuedGrinds").Return([]*entities.Grind{{
		ID: "g1",
		Participants: []entities.User{{ID: "u1"}},
	}}, nil)

	partRepo := new(mocks.MockParticipationRepository)
	partRepo.On("FindByUserAndGrind", "u1", "g1").Return(&entities.Participation{UserID: "u1", GrindID: "g1", TotalPenalty: 42}, nil)

	infoRepo := new(mocks.MockStripePaymentInfoRepository)
	infoRepo.On("FindByUserID", "u1").Return([]entities.StripePaymentInfo{{UserID: "u1", StripePaymentMethodID: "pm_1"}}, nil)

	svc := &StripePaymentService{
		grindRepo:             grindRepo,
		participationRepo:     partRepo,
		stripePaymentInfoRepo: infoRepo,
	}

	pending, err := svc.FindDuedPayments()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected one pending payment, got %d", len(pending))
	}
	if pending[0].PaymentAmount != 42 {
		t.Fatalf("expected payment amount 42, got %d", pending[0].PaymentAmount)
	}
	grindRepo.AssertExpectations(t)
	partRepo.AssertExpectations(t)
	infoRepo.AssertExpectations(t)
}
