package handler

import (
	"bytes"
	"encoding/json"
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

func TestRegisterGroupsV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	RegisterGroupsV1(router, &zerolog.Logger{}, store)

	assert.Equal(t, 6, len(router.Routes()))
}

func TestGetGroupsHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/groups", GetGroupsHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListGroups(gomock.Any()).Return([]*structs.Group{
			{
				Name:    "group1",
				Enabled: true,
			},
			{
				Name:    "group2",
				Enabled: false,
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/groups", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		assert.JSONEq(t, `{"groups":[{"name":"group1","enabled":true},{"name":"group2","enabled":false}]}`, rec.Body.String())
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
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&structs.Group{
			Name:    "group1",
			Enabled: true,
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/groups/group1", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		assert.JSONEq(t, `{"name":"group1","enabled":true}`, rec.Body.String())
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
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&structs.Group{
			Name:    "group1",
			Enabled: true,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &structs.Group{
			Name:    "group1",
			Enabled: false,
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
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&structs.Group{
			Name:    "group1",
			Enabled: true,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &structs.Group{
			Name:    "group1",
			Enabled: false,
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
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&structs.Group{
			Name:    "group1",
			Enabled: false,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &structs.Group{
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
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&structs.Group{
			Name:    "group1",
			Enabled: false,
		}, nil)
		store.EXPECT().SaveGroup(gomock.Any(), &structs.Group{
			Name:    "group1",
			Enabled: true,
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
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&structs.Group{
			Name:    "group1",
			Enabled: false,
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
		store.EXPECT().GetGroup(gomock.Any(), "group1").Return(&structs.Group{
			Name:    "group1",
			Enabled: false,
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
		store.EXPECT().SaveGroup(gomock.Any(), &structs.Group{
			Name:    "group1",
			Enabled: false,
		}).Return(nil)

		body, err := json.Marshal(&structs.Group{
			Name:    "group1",
			Enabled: false,
		})
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/groups", bytes.NewReader(body))
		router.ServeHTTP(rec, req)

		assert.Equal(t, 201, rec.Code)
		assert.JSONEq(t, `{"name":"group1","enabled":false}`, rec.Body.String())
	})

	t.Run("error on SaveGroup()", func(t *testing.T) {
		store.EXPECT().SaveGroup(gomock.Any(), &structs.Group{
			Name:    "group1",
			Enabled: false,
		}).Return(errors.New("error"))

		body, err := json.Marshal(&structs.Group{
			Name:    "group1",
			Enabled: false,
		})
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/groups", bytes.NewReader(body))
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("invalid body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/groups", bytes.NewReader([]byte("invalid")))
		router.ServeHTTP(rec, req)

		assert.Equal(t, 400, rec.Code)
		assert.JSONEq(t, `{"error":"invalid character 'i' looking for beginning of value"}`, rec.Body.String())
	})
}

func init() {
	gin.SetMode(gin.TestMode)
}
