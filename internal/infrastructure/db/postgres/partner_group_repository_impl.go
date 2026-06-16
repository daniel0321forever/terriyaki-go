package postgres

import (
	"context"

	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"gorm.io/gorm"
)

// partnerGroupMemberSchema is a minimal user schema used exclusively for the
// many2many join on partner_groups. It avoids importing UserSchema across files
// while keeping GORM happy with the group_members join table.
type partnerGroupMemberSchema struct {
	ID string `gorm:"primaryKey"`
}

func (partnerGroupMemberSchema) TableName() string { return "users" }

type PartnerGroupSchema struct {
	gorm.Model
	ID          string                     `json:"id" gorm:"primaryKey"`
	Name        string                     `json:"name" gorm:"not null"`
	InviteToken string                     `json:"invite_token" gorm:"uniqueIndex;not null"`
	OwnerID     string                     `json:"owner_id" gorm:"not null"`
	GrindID     string                     `json:"grind_id" gorm:"not null"`
	Members     []partnerGroupMemberSchema `gorm:"many2many:group_members;foreignKey:ID;joinForeignKey:PartnerGroupID;References:ID;joinReferences:UserID"`
}

func (PartnerGroupSchema) TableName() string { return "partner_groups" }

type GormPartnerGroupRepository struct {
	db *gorm.DB
}

func NewGormPartnerGroupRepository(db *gorm.DB) *GormPartnerGroupRepository {
	return &GormPartnerGroupRepository{db: db}
}

func partnerGroupSchemaToEntity(s *PartnerGroupSchema) *entities.PartnerGroup {
	members := make([]string, len(s.Members))
	for i, m := range s.Members {
		members[i] = m.ID
	}
	return &entities.PartnerGroup{
		ID:          s.ID,
		Name:        s.Name,
		InviteToken: s.InviteToken,
		OwnerID:     s.OwnerID,
		GrindID:     s.GrindID,
		Members:     members,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func (r *GormPartnerGroupRepository) Create(group *entities.PartnerGroup) error {
	ctx := context.Background()
	model := PartnerGroupSchema{
		ID:          group.ID,
		Name:        group.Name,
		InviteToken: group.InviteToken,
		OwnerID:     group.OwnerID,
		GrindID:     group.GrindID,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormPartnerGroupRepository) FindByID(id string) (*entities.PartnerGroup, error) {
	ctx := context.Background()
	var model PartnerGroupSchema
	if err := r.db.WithContext(ctx).Preload("Members").First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return partnerGroupSchemaToEntity(&model), nil
}

func (r *GormPartnerGroupRepository) FindByGrindID(grindID string) (*entities.PartnerGroup, error) {
	ctx := context.Background()
	var model PartnerGroupSchema
	if err := r.db.WithContext(ctx).Preload("Members").Where("grind_id = ?", grindID).First(&model).Error; err != nil {
		return nil, err
	}
	return partnerGroupSchemaToEntity(&model), nil
}

func (r *GormPartnerGroupRepository) FindByInviteToken(token string) (*entities.PartnerGroup, error) {
	ctx := context.Background()
	var model PartnerGroupSchema
	if err := r.db.WithContext(ctx).Preload("Members").Where("invite_token = ?", token).First(&model).Error; err != nil {
		return nil, err
	}
	return partnerGroupSchemaToEntity(&model), nil
}

func (r *GormPartnerGroupRepository) AddMember(groupID, userID string) error {
	ctx := context.Background()
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO group_members (partner_group_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		groupID, userID,
	).Error
}

func (r *GormPartnerGroupRepository) Update(group *entities.PartnerGroup) error {
	ctx := context.Background()
	model := PartnerGroupSchema{
		ID:          group.ID,
		Name:        group.Name,
		InviteToken: group.InviteToken,
		OwnerID:     group.OwnerID,
		GrindID:     group.GrindID,
	}
	return r.db.WithContext(ctx).Model(&PartnerGroupSchema{}).Where("id = ?", group.ID).Updates(&model).Error
}
