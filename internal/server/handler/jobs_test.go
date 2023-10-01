package handler

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/internal/server/store/mock"
	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterJobsV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	RegisterJobsV1(router, &zerolog.Logger{}, store)

	assert.Equal(t, 3, len(router.Routes()))
}

func TestGetJobsHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/jobs", GetJobsHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListJobs(gomock.Any()).Return([]*structs.Job{
			{
				ID:           "1",
				Name:         "job1",
				Organisation: "org1",
				Repository:   "repo1",
			},
			{
				ID:           "2",
				Name:         "job2",
				Organisation: "org1",
				Repository:   "repo1",
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/jobs", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		type response struct {
			Jobs []*structs.Job `storeon:"jobs"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(resp.Jobs))
	})

	t.Run("success with organisation query", func(t *testing.T) {
		store.EXPECT().ListJobs(gomock.Any()).Return([]*structs.Job{
			{
				ID:           "1",
				Name:         "job1",
				Organisation: "org1",
				Repository:   "repo1",
			},
			{
				ID:           "2",
				Name:         "job2",
				Organisation: "org2",
				Repository:   "repo1",
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/jobs?organisation=org1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		type response struct {
			Jobs []*structs.Job `storeon:"jobs"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(resp.Jobs))
	})

	t.Run("success with repository query", func(t *testing.T) {
		store.EXPECT().ListJobs(gomock.Any()).Return([]*structs.Job{
			{
				ID:           "1",
				Name:         "job1",
				Organisation: "org1",
				Repository:   "repo1",
			},
			{
				ID:           "2",
				Name:         "job2",
				Organisation: "org2",
				Repository:   "repo2",
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/jobs?repository=repo1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		type response struct {
			Jobs []*structs.Job `storeon:"jobs"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(resp.Jobs))
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().ListJobs(gomock.Any()).Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/jobs", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestGetJobHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/jobs/:id", GetJobHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetJob(gomock.Any(), "1").Return(&structs.Job{
			ID:           "1",
			Name:         "job1",
			Organisation: "org1",
			Repository:   "repo1",
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/jobs/1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		var job structs.Job
		err := json.Unmarshal(rec.Body.Bytes(), &job)
		assert.NoError(t, err)

		assert.Equal(t, "1", job.ID)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().GetJob(gomock.Any(), "1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/jobs/1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestDeleteJobHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.DELETE("/jobs/:id", DeleteJobHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().DeleteJob(gomock.Any(), "1").Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/jobs/1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().DeleteJob(gomock.Any(), "1").Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/jobs/1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}
