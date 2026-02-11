package services

import (
	"errors"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/mappers"
	"github.com/daniel0321forever/terriyaki-go/internal/cores/utils"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(request dto.CreateUserDTO) (*dto.UserDTO, error) {
	// Check if user already exists
	existing, _ := s.userRepo.FindByEmail(request.Email)
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	user, err := entities.NewUser(
		request.Username,
		request.Email,
		hashedPassword,
		request.Avatar,
	)
	if err != nil {
		return nil, err
	}

	// Save to database
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return mappers.UserToUserDTO(user), nil
}

func (s *UserService) GetUser(request dto.GetUserDTO) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return mappers.UserToUserDTO(user), nil
}

func (s *UserService) GetUserByEmail(request dto.GetUserByEmailDTO) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindByEmail(request.Email)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return mappers.UserToUserDTO(user), nil
}

func (s *UserService) DeleteUser(request dto.DeleteUserDTO) error {
	err := s.userRepo.Delete(request.UserID)
	return err
}

func (s *UserService) VerifyPassword(request dto.VerifyPasswordDTO) bool {
	return utils.VerifyPassword(request.UserID, request.Password)
}

func (s *UserService) UpdateUser(request dto.UpdateUserDTO) (*dto.UserDTO, error) {
	user, err := s.userRepo.FindById(request.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if request.Username != nil {
		user.Username = *request.Username
	}
	if request.Avatar != nil {
		user.Avatar = *request.Avatar
	}
	if request.DefaultPaymentMethodID != nil {
		user.DefaultPaymentMethodID = *request.DefaultPaymentMethodID
	}

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return mappers.UserToUserDTO(user), nil
}
