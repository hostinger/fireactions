package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/server/store/mock"
	"github.com/hostinger/fireactions/server/structs"
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

	assert.Equal(t, 5, len(router.Routes()))
}

func TestGetFlavorsHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/flavors", GetFlavorsHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListFlavors(gomock.Any()).Return([]*structs.Flavor{
			{
				Name:         "flavor1",
				Enabled:      true,
				DiskSizeGB:   10,
				MemorySizeMB: 1024,
				VCPUs:        2,
				Image:        "ubuntu-18.04",
			},
			{
				Name:         "flavor2",
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
		assert.JSONEq(t, `{"flavors":[{"name":"flavor1","enabled":true,"disk_size_gb":10,"memory_size_mb":1024,"vcpus":2,"image":"ubuntu-18.04"},{"name":"flavor2","enabled":true,"disk_size_gb":10,"memory_size_mb":1024,"vcpus":2,"image":"ubuntu-18.04"}]}`, rec.Body.String())
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
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&structs.Flavor{
			Name:         "flavor1",
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
		assert.JSONEq(t, `{"name":"flavor1","enabled":true,"disk_size_gb":10,"memory_size_mb":1024,"vcpus":2,"image":"ubuntu-18.04"}`, rec.Body.String())
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
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&structs.Flavor{
			Name:         "flavor1",
			Enabled:      true,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &structs.Flavor{
			Name:         "flavor1",
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
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&structs.Flavor{
			Name:         "flavor1",
			Enabled:      true,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &structs.Flavor{
			Name:         "flavor1",
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
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&structs.Flavor{
			Name:         "flavor1",
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &structs.Flavor{
			Name:         "flavor1",
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
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&structs.Flavor{
			Name:         "flavor1",
			Enabled:      false,
			DiskSizeGB:   10,
			MemorySizeMB: 1024,
			VCPUs:        2,
			Image:        "ubuntu-18.04",
		}, nil)

		store.EXPECT().SaveFlavor(gomock.Any(), &structs.Flavor{
			Name:         "flavor1",
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
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&structs.Flavor{
			Name:         "flavor1",
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
		store.EXPECT().GetFlavor(gomock.Any(), "flavor1").Return(&structs.Flavor{
			Name:         "flavor1",
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
