package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/internal/server/store/mock"
	"github.com/hostinger/fireactions/internal/server/structs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterRunnersV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	RegisterRunnersV1(router, &zerolog.Logger{}, store)

	assert.Equal(t, 2, len(router.Routes()))
}

func TestGetRunnersHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/runners", GetRunnersHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListRunners(gomock.Any()).Return([]*structs.Runner{
			{
				ID:           "1",
				Node:         nil,
				Name:         "runner1",
				Organisation: "org1",
				Group:        &structs.Group{Name: "group1", Enabled: true},
				Status:       structs.RunnerStatusPending,
				Labels:       "label1,label2",
				Flavor:       &structs.Flavor{Name: "flavor1", Enabled: true, DiskSizeGB: 10, MemorySizeMB: 1024, VCPUs: 2, Image: "ubuntu-18.04"},
				CreatedAt:    time.Unix(0, 0),
				UpdatedAt:    time.Unix(0, 0),
			},
			{
				ID:           "2",
				Node:         nil,
				Name:         "runner1",
				Organisation: "org1",
				Group:        &structs.Group{Name: "group1", Enabled: true},
				Status:       structs.RunnerStatusPending,
				Labels:       "label1,label2",
				Flavor:       &structs.Flavor{Name: "flavor1", Enabled: true, DiskSizeGB: 10, MemorySizeMB: 1024, VCPUs: 2, Image: "ubuntu-18.04"},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/runners", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		type response struct {
			Runners []*structs.Runner `json:"runners"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(resp.Runners))
	})

	t.Run("success with query organisation", func(t *testing.T) {
		store.EXPECT().ListRunners(gomock.Any()).Return([]*structs.Runner{
			{
				ID:           "1",
				Node:         nil,
				Name:         "runner1",
				Organisation: "org1",
				Group:        &structs.Group{Name: "group1", Enabled: true},
				Status:       structs.RunnerStatusPending,
				Labels:       "label1,label2",
				Flavor:       &structs.Flavor{Name: "flavor1", Enabled: true, DiskSizeGB: 10, MemorySizeMB: 1024, VCPUs: 2, Image: "ubuntu-18.04"},
				CreatedAt:    time.Unix(0, 0),
				UpdatedAt:    time.Unix(0, 0),
			},
			{
				ID:           "2",
				Node:         nil,
				Name:         "runner1",
				Organisation: "org2",
				Group:        &structs.Group{Name: "group1", Enabled: true},
				Status:       structs.RunnerStatusPending,
				Labels:       "label1,label2",
				Flavor:       &structs.Flavor{Name: "flavor1", Enabled: true, DiskSizeGB: 10, MemorySizeMB: 1024, VCPUs: 2, Image: "ubuntu-18.04"},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/runners?organisation=org1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		type response struct {
			Runners []*structs.Runner `json:"runners"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(resp.Runners))
	})

	t.Run("success with query group", func(t *testing.T) {
		store.EXPECT().ListRunners(gomock.Any()).Return([]*structs.Runner{
			{
				ID:           "1",
				Node:         nil,
				Name:         "runner1",
				Organisation: "org1",
				Group:        &structs.Group{Name: "group1", Enabled: true},
				Status:       structs.RunnerStatusPending,
				Labels:       "label1,label2",
				Flavor:       &structs.Flavor{Name: "flavor1", Enabled: true, DiskSizeGB: 10, MemorySizeMB: 1024, VCPUs: 2, Image: "ubuntu-18.04"},
				CreatedAt:    time.Unix(0, 0),
				UpdatedAt:    time.Unix(0, 0),
			},
			{
				ID:           "2",
				Node:         nil,
				Name:         "runner1",
				Organisation: "org2",
				Group:        &structs.Group{Name: "group2", Enabled: true},
				Status:       structs.RunnerStatusPending,
				Labels:       "label1,label2",
				Flavor:       &structs.Flavor{Name: "flavor1", Enabled: true, DiskSizeGB: 10, MemorySizeMB: 1024, VCPUs: 2, Image: "ubuntu-18.04"},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/runners?group=group1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		type response struct {
			Runners []*structs.Runner `json:"runners"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(resp.Runners))
	})

	t.Run("error on ListRunners()", func(t *testing.T) {
		store.EXPECT().ListRunners(gomock.Any()).Return(nil, assert.AnError)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/runners", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
	})
}

func TestGetRunnerHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/runnestore/:id", GetRunnerHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetRunner(gomock.Any(), "1").Return(&structs.Runner{
			ID:           "1",
			Node:         nil,
			Name:         "runner1",
			Organisation: "org1",
			Group:        &structs.Group{Name: "group1", Enabled: true},
			Status:       structs.RunnerStatusPending,
			Labels:       "label1,label2",
			Flavor:       &structs.Flavor{Name: "flavor1", Enabled: true, DiskSizeGB: 10, MemorySizeMB: 1024, VCPUs: 2, Image: "ubuntu-18.04"},
			CreatedAt:    time.Unix(0, 0),
			UpdatedAt:    time.Unix(0, 0),
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/runnestore/1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		var runner *structs.Runner
		err := json.Unmarshal(rec.Body.Bytes(), &runner)
		assert.NoError(t, err)

		assert.Equal(t, "1", runner.ID)
	})

	t.Run("error on GetRunner()", func(t *testing.T) {
		store.EXPECT().GetRunner(gomock.Any(), "1").Return(nil, assert.AnError)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/runnestore/1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
	})
}
