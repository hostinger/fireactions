package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterGroupsV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	RegisterGroupsV1(router, &zerolog.Logger{}, store)

	assert.Equal(t, 7, len(router.Routes()))
}

func TestGetGroupsHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/groups", GetGroupsHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListGroups(gomock.Any()).Return([]*models.Group{
			{
				Name:      "group1",
				Enabled:   true,
				IsDefault: true,
			},
			{
				Name:      "group2",
				Enabled:   false,
				IsDefault: false,
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/groups", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		type Root struct {
			Groups []*api.Group `json:"groups"`
		}

		var root Root
		err := json.Unmarshal(rec.Body.Bytes(), &root)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(root.Groups))
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().ListGroups(gomock.Any()).Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", "/groups", nil))

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestGetGroupHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/groups/:name", GetGroupHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   true,
			IsDefault: true,
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/groups/group1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		var group api.Group
		err := json.Unmarshal(rec.Body.Bytes(), &group)
		assert.NoError(t, err)

		assert.Equal(t, "group1", group.Name)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/groups/group1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestDisableGroupHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.PATCH("/groups/:name/disable", DisableGroupHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   true,
			IsDefault: true,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: true,
		}).Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/disable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/disable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on SaveGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   true,
			IsDefault: true,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: true,
		}).Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/disable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestEnableGroupHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.PATCH("/groups/:name/enable", EnableGroupHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:    "group1",
			Enabled: false,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &models.Group{
			Name:    "group1",
			Enabled: true,
		}).Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/enable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/enable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on SaveGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &models.Group{
			Name:      "group1",
			Enabled:   true,
			IsDefault: false,
		}).Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/enable", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestDeleteGroupHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.DELETE("/groups/:name", DeleteGroupHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		}, nil)
		store.EXPECT().DeleteGroup(gomock.Any(), "group1").Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/groups/group1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/groups/group1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on DeleteGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		}, nil)
		store.EXPECT().DeleteGroup(gomock.Any(), "group1").Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/groups/group1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestCreateGroupHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.POST("/groups", CreateGroupHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().SaveGroup(gomock.Any(), &models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		}).Return(nil)

		body, err := json.Marshal(&models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		})
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/groups", bytes.NewReader(body))
		router.ServeHTTP(rec, req)

		assert.Equal(t, 201, rec.Code)

		var group api.Group
		err = json.Unmarshal(rec.Body.Bytes(), &group)
		assert.NoError(t, err)

		assert.Equal(t, "group1", group.Name)
	})

	t.Run("error on SaveGroup()", func(t *testing.T) {
		store.EXPECT().SaveGroup(gomock.Any(), &models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		}).Return(errors.New("error"))

		body, err := json.Marshal(&models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		})
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/groups", bytes.NewReader(body))
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on BindJSON()", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/groups", bytes.NewReader([]byte("invalid")))
		router.ServeHTTP(rec, req)

		assert.Equal(t, 400, rec.Code)
		assert.JSONEq(t, `{"error":"invalid character 'i' looking for beginning of value"}`, rec.Body.String())
	})
}

func TestSetDefaultGroupHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.PATCH("/groups/:name/default", SetDefaultGroupHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		}, nil)
		store.EXPECT().SetDefaultGroup(gomock.Any(), "group1").Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/default", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/default", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on SetDefaultGroup()", func(t *testing.T) {
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&models.Group{
			Name:      "group1",
			Enabled:   false,
			IsDefault: false,
		}, nil)
		store.EXPECT().SetDefaultGroup(gomock.Any(), "group1").Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/groups/group1/default", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func init() {
	gin.SetMode(gin.TestMode)
}
