package services

import (
	"fmt"
	"os"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/repositories"
	"github.com/golang-jwt/jwt/v5"
)

// PartnerGroupService manages partner group creation and the shareable invite flow.
type PartnerGroupService struct {
	partnerGroupRepo repositories.PartnerGroupRepository
}

// NewPartnerGroupService constructs a PartnerGroupService with the given repository.
func NewPartnerGroupService(partnerGroupRepo repositories.PartnerGroupRepository) *PartnerGroupService {
	return &PartnerGroupService{
		partnerGroupRepo: partnerGroupRepo,
	}
}

// CreateGroup creates a new PartnerGroup for the given grind and owner.
func (s *PartnerGroupService) CreateGroup(grindID, ownerID, name string) (*entities.PartnerGroup, error) {
	group, err := entities.NewPartnerGroup(grindID, ownerID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create partner group entity: %w", err)
	}

	if err := s.partnerGroupRepo.Create(group); err != nil {
		return nil, fmt.Errorf("failed to persist partner group: %w", err)
	}

	return group, nil
}

// GetGroup retrieves a partner group by its ID.
func (s *PartnerGroupService) GetGroup(groupID string) (*entities.PartnerGroup, error) {
	return s.partnerGroupRepo.FindByID(groupID)
}

// GenerateInviteToken generates a JWT invite token for the given group.
// Returns ErrForbidden if callerID is not the group owner.
// Claims: sub=groupID, type="partner_group_invite", iss="habitat", exp=7 days.
func (s *PartnerGroupService) GenerateInviteToken(groupID, callerID string) (string, error) {
	group, err := s.partnerGroupRepo.FindByID(groupID)
	if err != nil {
		return "", fmt.Errorf("failed to find partner group: %w", err)
	}

	if group.OwnerID != callerID {
		return "", config.ErrForbidden
	}

	claims := jwt.MapClaims{
		"sub":  groupID,
		"type": "partner_group_invite",
		"iss":  "habitat",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to sign invite token: %w", err)
	}

	return signed, nil
}

// JoinGroup validates a JWT invite token and adds userID to the group.
// Returns an error containing "token is expired" for expired tokens.
func (s *PartnerGroupService) JoinGroup(tokenString, userID string) (*entities.PartnerGroup, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid invite token")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "partner_group_invite" {
		return nil, fmt.Errorf("invalid token type: expected partner_group_invite, got %s", tokenType)
	}

	groupID, ok := claims["sub"].(string)
	if !ok || groupID == "" {
		return nil, fmt.Errorf("invalid token: missing group ID")
	}

	if _, err := s.partnerGroupRepo.FindByID(groupID); err != nil {
		return nil, fmt.Errorf("failed to find partner group: %w", err)
	}

	if err := s.partnerGroupRepo.AddMember(groupID, userID); err != nil {
		return nil, fmt.Errorf("failed to add member to group: %w", err)
	}

	updated, err := s.partnerGroupRepo.FindByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated group: %w", err)
	}

	return updated, nil
}
