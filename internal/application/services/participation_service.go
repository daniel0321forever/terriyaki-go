package services

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
)

type ParticipationService struct {
	participationRepo repositories.ParticipationRepository
	taskRepo          repositories.TaskRepository
}

func NewParticipationService(
	participationRepo repositories.ParticipationRepository,
	taskRepo repositories.TaskRepository,
) *ParticipationService {
	return &ParticipationService{
		participationRepo: participationRepo,
		taskRepo:          taskRepo,
	}
}

// GetParticipation retrieves a participation record by id
func (s *ParticipationService) GetParticipation(request dto.GetParticipation) (*dto.ParticipationDTO, error) {
	participation, err := s.participationRepo.FindByParticipationId(request.ParticipationID)
	if err != nil {
		return nil, config.ErrParticipationNotFound
	}

	return mappers.ParticipationToParticipationDTO(participation), nil
}

// GetParticipationByUserAndGrind retrieves a participation record by user and grind
func (s *ParticipationService) GetParticipationByUserAndGrind(request dto.GetParticipationByUserAndGrindDTO) (*dto.ParticipationDTO, error) {
	participation, err := s.participationRepo.FindByUserAndGrind(request.UserID, request.GrindID)
	if err != nil {
		return nil, config.ErrParticipationNotFound
	}

	return mappers.ParticipationToParticipationDTO(participation), nil
}

func (s *ParticipationService) UpdateParticipation(request dto.UpdateAddParticipationDTO) (*dto.ParticipationDTO, error) {
	participation, err := s.participationRepo.FindByParticipationId(request.ParticipationID)
	if err != nil {
		return nil, config.ErrParticipationNotFound
	}
	participation.MissedDays = request.MissedDays
	participation.TotalPenalty = request.TotalPenalty
	participation.Quitted = request.Quitted
	participation.QuittedAt = *request.QuittedAt
	err = s.participationRepo.Update(participation)
	if err != nil {
		return nil, config.ErrGrindUpdateFailed
	}
	return mappers.ParticipationToParticipationDTO(participation), nil
}
