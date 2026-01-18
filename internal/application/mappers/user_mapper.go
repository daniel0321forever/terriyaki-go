package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// Converts User entity to UserDTO
func UserToUserDTO(user *entities.User) *dto.UserDTO {
	return &dto.UserDTO{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Avatar:         user.Avatar,
		HashedPassword: user.HashedPassword,
	}
}
