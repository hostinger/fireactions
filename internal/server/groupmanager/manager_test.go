package groupmanager

import (
	"testing"

	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	gm := New(&zerolog.Logger{})

	assert.NotNil(t, gm)
}

func TestGroupManager_GetGroup(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroup(&structs.Group{Name: "test", Enabled: true})
	assert.NoError(t, err)

	group, err := gm.GetGroup("test")
	assert.NoError(t, err)

	assert.Equal(t, "test", group.Name)
	assert.Equal(t, true, group.Enabled)
}

func TestGroupManager_GetGroup_NotFound(t *testing.T) {
	gm := New(&zerolog.Logger{})

	_, err := gm.GetGroup("test")
	assert.ErrorIs(t, err, ErrGroupNotFound)
}

func TestGroupManager_ListGroups(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroup(&structs.Group{Name: "test", Enabled: true})
	assert.NoError(t, err)

	groups, err := gm.ListGroups()
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, groups, 1)
}

func TestGroupManager_AddGroup(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroup(&structs.Group{Name: "test", Enabled: true})
	assert.NoError(t, err)

	group, err := gm.GetGroup("test")
	assert.NoError(t, err)

	assert.Equal(t, "test", group.Name)
	assert.Equal(t, true, group.Enabled)
}

func TestGroupManager_AddGroup_Exists(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroup(&structs.Group{Name: "test", Enabled: true})
	assert.NoError(t, err)

	err = gm.AddGroup(&structs.Group{Name: "test", Enabled: true})
	assert.ErrorIs(t, err, ErrGroupExists)
}

func TestGroupManager_AddGroups(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroups([]*structs.Group{
		{Name: "test1", Enabled: true},
		{Name: "test2", Enabled: true},
	}...)
	assert.NoError(t, err)

	group, err := gm.GetGroup("test1")
	assert.NoError(t, err)

	assert.Equal(t, "test1", group.Name)
	assert.Equal(t, true, group.Enabled)

	group, err = gm.GetGroup("test2")
	assert.NoError(t, err)

	assert.Equal(t, "test2", group.Name)
	assert.Equal(t, true, group.Enabled)
}

func TestGroupManager_AddGroups_Exists(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroups([]*structs.Group{
		{Name: "test1", Enabled: true},
		{Name: "test2", Enabled: true},
	}...)
	assert.NoError(t, err)

	err = gm.AddGroups([]*structs.Group{
		{Name: "test1", Enabled: true},
		{Name: "test2", Enabled: true},
	}...)
	assert.ErrorIs(t, err, ErrGroupExists)
}

func TestGroupManager_DisableGroup(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroup(&structs.Group{Name: "test", Enabled: true})
	assert.NoError(t, err)

	err = gm.DisableGroup("test")
	assert.NoError(t, err)

	group, err := gm.GetGroup("test")
	assert.NoError(t, err)

	assert.False(t, group.Enabled)
}

func TestGroupManager_EnableGroup(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.AddGroup(&structs.Group{Name: "test", Enabled: false})
	assert.NoError(t, err)

	err = gm.EnableGroup("test")
	assert.NoError(t, err)

	group, err := gm.GetGroup("test")
	assert.NoError(t, err)

	assert.True(t, group.Enabled)
}

func TestGroupManager_EnableGroup_NotFound(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.EnableGroup("test")
	assert.ErrorIs(t, err, ErrGroupNotFound)
}

func TestGroupManager_DisableGroup_NotFound(t *testing.T) {
	gm := New(&zerolog.Logger{})

	err := gm.DisableGroup("test")
	assert.ErrorIs(t, err, ErrGroupNotFound)
}
