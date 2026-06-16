package services

import (
	"os"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/cores/config"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_PartnerGroupService_CreateGroup_Success(t *testing.T) {
	t.Parallel()

	partnerGroupRepo := new(mocks.MockPartnerGroupRepository)
	partnerGroupRepo.On("Create", mock.MatchedBy(func(g *entities.PartnerGroup) bool {
		return g.GrindID == "grind-1" && g.OwnerID == "user-1" && g.Name == "Study Group"
	})).Return(nil)

	svc := NewPartnerGroupService(partnerGroupRepo)

	group, err := svc.CreateGroup("grind-1", "user-1", "Study Group")
	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, "grind-1", group.GrindID)
	assert.Equal(t, "user-1", group.OwnerID)
	assert.Equal(t, "Study Group", group.Name)

	partnerGroupRepo.AssertExpectations(t)
}

func Test_PartnerGroupService_GenerateInviteToken_Forbidden(t *testing.T) {
	t.Parallel()

	partnerGroupRepo := new(mocks.MockPartnerGroupRepository)
	group := &entities.PartnerGroup{
		ID:      "group-1",
		OwnerID: "user-1",
		GrindID: "grind-1",
		Name:    "Study Group",
	}
	partnerGroupRepo.On("FindByID", "group-1").Return(group, nil)

	svc := NewPartnerGroupService(partnerGroupRepo)

	// caller is "user-2", not the owner "user-1"
	_, err := svc.GenerateInviteToken("group-1", "user-2")
	assert.Error(t, err)
	assert.ErrorIs(t, err, config.ErrForbidden)

	partnerGroupRepo.AssertExpectations(t)
}

func Test_PartnerGroupService_GenerateInviteToken_Success(t *testing.T) {
	// Not parallel — uses os.Setenv which is not goroutine-safe
	require := assert.New(t)
	require.NoError(os.Setenv("JWT_SECRET", "test-secret"))
	defer func() { _ = os.Unsetenv("JWT_SECRET") }()

	partnerGroupRepo := new(mocks.MockPartnerGroupRepository)
	group := &entities.PartnerGroup{
		ID:      "group-1",
		OwnerID: "user-1",
		GrindID: "grind-1",
		Name:    "Study Group",
	}
	partnerGroupRepo.On("FindByID", "group-1").Return(group, nil)

	svc := NewPartnerGroupService(partnerGroupRepo)

	token, err := svc.GenerateInviteToken("group-1", "user-1")
	require.NoError(err)
	require.NotEmpty(token)

	// Verify the token contains the right claims
	parsed, parseErr := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	require.NoError(parseErr)
	require.True(parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(ok)
	require.Equal("group-1", claims["sub"])
	require.Equal("partner_group_invite", claims["type"])
	require.Equal("habitat", claims["iss"])

	partnerGroupRepo.AssertExpectations(t)
}

func Test_PartnerGroupService_JoinGroup_Success(t *testing.T) {
	// Not parallel — uses os.Setenv which is not goroutine-safe
	require := assert.New(t)
	require.NoError(os.Setenv("JWT_SECRET", "test-secret"))
	defer func() { _ = os.Unsetenv("JWT_SECRET") }()

	// Generate a valid token
	tokenClaims := jwt.MapClaims{
		"sub":  "group-1",
		"type": "partner_group_invite",
		"iss":  "habitat",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenStr, err := tokenObj.SignedString([]byte("test-secret"))
	require.NoError(err)

	group := &entities.PartnerGroup{
		ID:      "group-1",
		OwnerID: "user-1",
		GrindID: "grind-1",
		Name:    "Study Group",
		Members: []string{"user-1"},
	}
	updatedGroup := &entities.PartnerGroup{
		ID:      "group-1",
		OwnerID: "user-1",
		GrindID: "grind-1",
		Name:    "Study Group",
		Members: []string{"user-1", "user-2"},
	}

	partnerGroupRepo := new(mocks.MockPartnerGroupRepository)
	partnerGroupRepo.On("FindByID", "group-1").Return(group, nil).Once()
	partnerGroupRepo.On("AddMember", "group-1", "user-2").Return(nil)
	partnerGroupRepo.On("FindByID", "group-1").Return(updatedGroup, nil).Once()

	svc := NewPartnerGroupService(partnerGroupRepo)

	result, joinErr := svc.JoinGroup(tokenStr, "user-2")
	require.NoError(joinErr)
	require.NotNil(result)
	require.Contains(result.Members, "user-2")

	partnerGroupRepo.AssertExpectations(t)
}

func Test_PartnerGroupService_JoinGroup_ExpiredToken(t *testing.T) {
	// Not parallel — uses os.Setenv which is not goroutine-safe
	require := assert.New(t)
	require.NoError(os.Setenv("JWT_SECRET", "test-secret"))
	defer func() { _ = os.Unsetenv("JWT_SECRET") }()

	// Generate an expired token
	tokenClaims := jwt.MapClaims{
		"sub":  "group-1",
		"type": "partner_group_invite",
		"iss":  "habitat",
		"exp":  time.Now().Add(-1 * time.Second).Unix(), // expired
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenStr, err := tokenObj.SignedString([]byte("test-secret"))
	require.NoError(err)

	partnerGroupRepo := new(mocks.MockPartnerGroupRepository)

	svc := NewPartnerGroupService(partnerGroupRepo)

	_, joinErr := svc.JoinGroup(tokenStr, "user-2")
	require.Error(joinErr)
	require.Contains(joinErr.Error(), "token is expired")

	partnerGroupRepo.AssertExpectations(t)
}
