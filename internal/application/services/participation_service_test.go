package services

import (
	"errors"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
	"github.com/stretchr/testify/mock"
)

func TestParticipationServiceGetParticipation_NotFoundMapped(t *testing.T) {
	t.Parallel()

	partRepo := new(mocks.MockParticipationRepository)
	partRepo.On("FindByParticipationId", "p1").Return(nil, errors.New("missing"))

	svc := NewParticipationService(partRepo)

	_, err := svc.GetParticipation(dto.GetParticipation{ParticipationID: "p1"})
	if !errors.Is(err, config.ErrParticipationNotFound) {
		t.Fatalf("expected ErrParticipationNotFound, got %v", err)
	}
}

func TestParticipationServiceUpdateParticipation_Success(t *testing.T) {
	t.Parallel()

	q := time.Now().UTC()
	model := &entities.Participation{ID: "p1", UserID: "u1", GrindID: "g1"}

	partRepo := new(mocks.MockParticipationRepository)

	partRepo.On("FindByParticipationId", "p1").Return(model, nil)
	partRepo.On("Update", mock.MatchedBy(func(p *entities.Participation) bool {
		return p.ID == "p1" && p.MissedDays == 3 && p.TotalPenalty == 120 && p.Quitted
	})).Return(nil)

	svc := NewParticipationService(partRepo)

	res, err := svc.UpdateParticipation(dto.UpdateAddParticipationDTO{
		ParticipationID: "p1",
		MissedDays:      3,
		TotalPenalty:    120,
		Quitted:         true,
		QuittedAt:       &q,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.MissedDays != 3 || res.TotalPenalty != 120 || !res.Quitted {
		t.Fatalf("expected updated participation fields")
	}
	partRepo.AssertExpectations(t)
}
