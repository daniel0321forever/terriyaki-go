package dto

// Input DTOs
type CreateUserDTO struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Avatar   string `json:"avatar,omitempty"`
}

type UpdateUserDTO struct {
	UserID                 string  `json:"userID"`
	Username               *string `json:"username"`
	Avatar                 *string `json:"avatar"`
	DefaultPaymentMethodID *string `json:"defaultPaymentMethodID"`
}

type GetUserDTO struct {
	UserID string
}

type GetUserByEmailDTO struct {
	Email string
}

type DeleteUserDTO struct {
	UserID string
}

type VerifyPasswordDTO struct {
	UserID   string
	Password string
}

// Output DTOs
type UserDTO struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Avatar         string `json:"avatar"`
	HashedPassword string `json:"password"`
}
