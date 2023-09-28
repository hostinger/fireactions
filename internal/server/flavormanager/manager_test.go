package flavormanager

import (
	"testing"

	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	fm := New(&zerolog.Logger{})

	assert.NotNil(t, fm)
}

func TestNew_WithDefaultFlavor(t *testing.T) {
	fm := New(&zerolog.Logger{}, WithDefaultFlavor(&structs.Flavor{Name: "test", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50}))

	assert.Equal(t, "test", fm.GetDefaultFlavor().Name)
	assert.Equal(t, int64(1), fm.GetDefaultFlavor().VCPUs)
	assert.Equal(t, int64(1024), fm.GetDefaultFlavor().MemorySizeMB)
	assert.Equal(t, int64(50), fm.GetDefaultFlavor().DiskSizeGB)
}

func TestFlavorManager_GetDefaultFlavor(t *testing.T) {
	fm := New(&zerolog.Logger{})

	assert.NotNil(t, fm.GetDefaultFlavor())
}

func TestFlavorManager_GetFlavor(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.AddFlavor(&structs.Flavor{Name: "test", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50})
	assert.NoError(t, err)

	flavor, err := fm.GetFlavor("test")
	assert.NoError(t, err)

	assert.Equal(t, "test", flavor.Name)
	assert.Equal(t, int64(1), flavor.VCPUs)
	assert.Equal(t, int64(1024), flavor.MemorySizeMB)
	assert.Equal(t, int64(50), flavor.DiskSizeGB)
}

func TestFlavorManager_GetFlavor_NotFound(t *testing.T) {
	fm := New(&zerolog.Logger{})

	_, err := fm.GetFlavor("test")
	assert.ErrorIs(t, err, ErrFlavorNotFound)
}

func TestFlavorManager_ListFlavors(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.AddFlavor(&structs.Flavor{Name: "test", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50})
	assert.NoError(t, err)

	flavors, err := fm.ListFlavors()
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, flavors, 1)
}

func TestFlavorManager_AddFlavor(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.AddFlavor(&structs.Flavor{Name: "test", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50})
	assert.NoError(t, err)

	flavor, err := fm.GetFlavor("test")
	assert.NoError(t, err)

	assert.Equal(t, "test", flavor.Name)
	assert.Equal(t, int64(1), flavor.VCPUs)
	assert.Equal(t, int64(1024), flavor.MemorySizeMB)
	assert.Equal(t, int64(50), flavor.DiskSizeGB)
}

func TestFlavorManager_AddFlavors(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.AddFlavors([]*structs.Flavor{
		{Name: "test1", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50},
		{Name: "test2", VCPUs: 2, MemorySizeMB: 2048, DiskSizeGB: 100},
	}...)
	assert.NoError(t, err)

	flavor, err := fm.GetFlavor("test1")
	assert.NoError(t, err)

	assert.Equal(t, "test1", flavor.Name)
	assert.Equal(t, int64(1), flavor.VCPUs)
	assert.Equal(t, int64(1024), flavor.MemorySizeMB)
	assert.Equal(t, int64(50), flavor.DiskSizeGB)

	flavor, err = fm.GetFlavor("test2")
	assert.NoError(t, err)

	assert.Equal(t, "test2", flavor.Name)
	assert.Equal(t, int64(2), flavor.VCPUs)
	assert.Equal(t, int64(2048), flavor.MemorySizeMB)
	assert.Equal(t, int64(100), flavor.DiskSizeGB)
}

func TestFlavorManager_AddFlavors_AlreadyExists(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.AddFlavors([]*structs.Flavor{
		{Name: "test1", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50},
		{Name: "test2", VCPUs: 2, MemorySizeMB: 2048, DiskSizeGB: 100},
	}...)
	assert.NoError(t, err)

	err = fm.AddFlavors([]*structs.Flavor{
		{Name: "test1", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50},
		{Name: "test2", VCPUs: 2, MemorySizeMB: 2048, DiskSizeGB: 100},
	}...)
	assert.ErrorIs(t, err, ErrFlavorExists)
}

func TestFlavorManager_AddFlavor_AlreadyExists(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.AddFlavor(&structs.Flavor{Name: "test", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50})
	assert.NoError(t, err)

	err = fm.AddFlavor(&structs.Flavor{Name: "test", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50})
	assert.ErrorIs(t, err, ErrFlavorExists)
}

func TestFlavorManager_SetDefaultFlavor(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.AddFlavor(&structs.Flavor{Name: "test", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50})
	assert.NoError(t, err)

	err = fm.SetDefaultFlavor("test")
	assert.NoError(t, err)

	assert.Equal(t, "test", fm.GetDefaultFlavor().Name)
}

func TestFlavorManager_SetDefaultFlavor_NotFound(t *testing.T) {
	fm := New(&zerolog.Logger{})

	err := fm.SetDefaultFlavor("test")
	assert.ErrorIs(t, err, ErrFlavorNotFound)
}
