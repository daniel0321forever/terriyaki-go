package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewPartnerGroup_ValidInputs(t *testing.T) {
	group, err := NewPartnerGroup("grind-1", "user-1", "Study Group")
	require.NoError(t, err)
	require.NotNil(t, group)

	assert.NotEmpty(t, group.ID)
	assert.Equal(t, "grind-1", group.GrindID)
	assert.Equal(t, "user-1", group.OwnerID)
	assert.Equal(t, "Study Group", group.Name)
	assert.Greater(t, len(group.InviteToken), 0, "InviteToken should be non-empty")
	assert.Contains(t, group.Members, "user-1", "Members should contain OwnerID")
}

func Test_NewPartnerGroup_EmptyGrindID(t *testing.T) {
	group, err := NewPartnerGroup("", "user-1", "Study Group")
	require.Error(t, err)
	assert.Nil(t, group)
	assert.Equal(t, "grindID cannot be empty", err.Error())
}

func Test_NewPartnerGroup_EmptyOwnerID(t *testing.T) {
	group, err := NewPartnerGroup("grind-1", "", "Study Group")
	require.Error(t, err)
	assert.Nil(t, group)
	assert.Equal(t, "ownerID cannot be empty", err.Error())
}
