package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewHabitTask_ValidInputs(t *testing.T) {
	now := time.Now()
	task, err := NewHabitTask("user-1", "grind-1", now)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.NotEmpty(t, task.ID, "ID should be a non-empty UUID")
	assert.Equal(t, "generic", task.TaskType)
	assert.Equal(t, "user-1", task.UserID)
	assert.Equal(t, "grind-1", task.GrindID)
	assert.False(t, task.Completed)
	assert.Nil(t, task.Metadata)
}

func Test_NewHabitTask_EmptyUserID(t *testing.T) {
	now := time.Now()
	task, err := NewHabitTask("", "grind-1", now)
	require.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "userID cannot be empty", err.Error())
}

func Test_NewHabitTask_EmptyGrindID(t *testing.T) {
	now := time.Now()
	task, err := NewHabitTask("user-1", "", now)
	require.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "grindID cannot be empty", err.Error())
}
