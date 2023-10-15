package handler

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterFlavorsV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	RegisterFlavorsV1(router, &zerolog.Logger{}, store)

	assert.Equal(t, 6, len(router.Routes()))
}

func TestGetFlavorsHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/flavors", GetFlavorsHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListFlavors(gomock.Any()).Return([]*models.Flavor{
			{
				Name:         "flavor1",
				IsDefault:    true,
				Enabled:      true,
				DiskSizeGB:   10,
				MemorySizeMB: 1024,
				VCPUs:        2,
				Image:        "ubuntu-18.04",
			},
			{
				Name:         "flavor2",
				IsDefault:    false,
				Enabled:      true,
				DiskSizeGB:   10,
				MemorySizeMB: 1024,
				VCPUs:        2,
				Image:        "ubuntu-18.04",
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/flavors", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		type Root struct {
			Flavors []*models.Flavor `json:"flavors"`
		}

		var root Root
		err := json.Unmarshal(rec.Body.Bytes(), &root)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(root.Flavors))
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().ListFlavors(gomock.Any()).Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/flavors", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestGetFlavorHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/flavors/:name", GetFlavorHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      true,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		var flavor models.Flavor
		err := json.Unmarshal(rec.Body.Bytes(), &flavor)
		assert.NoError(t, err)

		assert.Equal(t, "flavor1", flavor.Name)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestDisableFlavorHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.DELETE("/flavors/:name", DisableFlavorHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      true,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}).Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on SaveFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      true,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}).Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestEnableFlavorHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.PATCH("/flavors/:name/enable", EnableFlavorHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      true,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}).Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/flavors/flavor1/enable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/flavors/flavor1/enable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on SaveFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      true,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}).Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/flavors/flavor1/enable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestDeleteFlavorHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.DELETE("/flavors/:name", DeleteFlavorHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().DeleteFlavor(gomock.Any(), "flavor1").Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
	})

	t.Run("error on DeleteFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    true,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().DeleteFlavor(gomock.Any(), "flavor1").Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/flavors/flavor1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
	})
}

func TestSetDefaultFlavorHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.PATCH("/flavors/:name/default", SetDefaultFlavorHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    false,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SetDefaultFlavor(gomock.Any(), "flavor1").Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/flavors/flavor1/default", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/flavors/flavor1/default", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
	})

	t.Run("error on SetDefaultFlavor()", func(t *testing.T) {
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&models.Flavor{
			Name:         "flavor1",
			IsDefault:    false,
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SetDefaultFlavor(gomock.Any(), "flavor1").Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/flavors/flavor1/default", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
	})
}
