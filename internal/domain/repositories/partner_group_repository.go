package repositories

import (
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// PartnerGroupRepository defines persistence operations for PartnerGroup domain entities.
type PartnerGroupRepository interface {
	Create(group *entities.PartnerGroup) error
	FindByID(id string) (*entities.PartnerGroup, error)
	FindByGrindID(grindID string) (*entities.PartnerGroup, error)
	FindByInviteToken(token string) (*entities.PartnerGroup, error)
	AddMember(groupID, userID string) error
	Update(group *entities.PartnerGroup) error
}
