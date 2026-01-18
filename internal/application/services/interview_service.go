package services

import (
	"encoding/json"
	"errors"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"gorm.io/datatypes"
)

type InterviewService struct {
	sessionRepo repositories.InterviewSessionRepository
}

func NewInterviewService(sessionRepo repositories.InterviewSessionRepository) *InterviewService {
	return &InterviewService{
		sessionRepo: sessionRepo,
	}
}

func (s *InterviewService) CreateSession(request dto.CreateInterviewSessionDTO) (*dto.InterviewSessionDTO, error) {
	// Create session entity using constructor (if it exists) or directly
	session, err := entities.NewInterviewSession(request.UserID, request.TaskID)
	if err != nil {
		return nil, err
	}

	err = s.sessionRepo.Create(session)
	if err != nil {
		return nil, err
	}

	return mappers.InterviewSessionToInterviewSessionDTO(session), nil
}

func (s *InterviewService) GetSession(request dto.GetInterviewSessionDTO) (*dto.InterviewSessionDTO, error) {
	session, err := s.sessionRepo.FindByID(request.SessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}
	return mappers.InterviewSessionToInterviewSessionDTO(session), nil
}

func (s *InterviewService) UpdateSession(request dto.UpdateInterviewSessionDTO) (*dto.InterviewSessionDTO, error) {
	session, err := s.sessionRepo.FindByID(request.SessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	// Update fields if provided
	if request.Status != nil {
		session.Status = *request.Status
	}
	if request.ConversationHistory != nil {
		// Convert interface{} to datatypes.JSON
		historyBytes, err := json.Marshal(request.ConversationHistory)
		if err != nil {
			return nil, err
		}
		session.ConversationHistory = datatypes.JSON(historyBytes)
	}
	if request.EndedAt != nil {
		session.EndedAt = request.EndedAt
	}

	err = s.sessionRepo.Update(session)
	if err != nil {
		return nil, err
	}

	return mappers.InterviewSessionToInterviewSessionDTO(session), nil
}

