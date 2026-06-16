package entities

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
)

// PartnerGroup represents a collective accountability group tied to a Grind (per D-04, D-05).
// Members is a slice of user IDs. InviteToken is a cryptographically random token used to
// join the group and must be treated as a secret (log masking required in controllers).
type PartnerGroup struct {
	ID          string
	Name        string
	InviteToken string
	OwnerID     string
	GrindID     string
	Members     []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewPartnerGroup creates a PartnerGroup with a generated UUID, a cryptographically
// random 32-byte URL-safe base64 InviteToken, and Members pre-populated with ownerID.
// grindID and ownerID must be non-empty.
func NewPartnerGroup(grindID, ownerID, name string) (*PartnerGroup, error) {
	if grindID == "" {
		return nil, errors.New("grindID cannot be empty")
	}
	if ownerID == "" {
		return nil, errors.New("ownerID cannot be empty")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.New("failed to generate invite token")
	}
	inviteToken := base64.RawURLEncoding.EncodeToString(tokenBytes)

	now := time.Now().UTC()
	return &PartnerGroup{
		ID:          uuid.New().String(),
		Name:        name,
		InviteToken: inviteToken,
		OwnerID:     ownerID,
		GrindID:     grindID,
		Members:     []string{ownerID},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
