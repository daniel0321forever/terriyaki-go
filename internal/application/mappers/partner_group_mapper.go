package mappers

import (
	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
)

// BuildPartnerGroupDTO constructs a PartnerGroupDTO from a PartnerGroup entity.
func BuildPartnerGroupDTO(group *entities.PartnerGroup) *dto.PartnerGroupDTO {
	members := group.Members
	if members == nil {
		members = []string{}
	}
	return &dto.PartnerGroupDTO{
		ID:        group.ID,
		Name:      group.Name,
		GrindID:   group.GrindID,
		OwnerID:   group.OwnerID,
		Members:   members,
		CreatedAt: group.CreatedAt,
	}
}
