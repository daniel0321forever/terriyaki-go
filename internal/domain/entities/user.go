package entities

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

/** Represents a user account **/
type User struct {
	ID                     string
	Username               string
	Email                  string
	Avatar                 string
	HashedPassword         string
	StripeCustomerID       string
	DefaultPaymentMethodID string
}

/** Constructor in factory pattern
 * @param username - the username
 * @param email - the email address
 * @param hashedPassword - the hashed password (hashing should be done in service layer)
 * @param avatar - the avatar URL (optional, can be empty)
 * @return the created user
 */
func NewUser(username, email, hashedPassword, avatar string) (*User, error) {
	// Validation
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username cannot be empty")
	}
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email cannot be empty")
	}
	if !strings.Contains(email, "@") {
		return nil, errors.New("invalid email format")
	}
	if strings.TrimSpace(hashedPassword) == "" {
		return nil, errors.New("password cannot be empty")
	}

	return &User{
		ID:             uuid.New().String(),
		Username:       strings.TrimSpace(username),
		Email:          strings.ToLower(strings.TrimSpace(email)),
		Avatar:         strings.TrimSpace(avatar),
		HashedPassword: hashedPassword,
	}, nil
}
