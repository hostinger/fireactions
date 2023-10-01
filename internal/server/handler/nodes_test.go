package handler

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/internal/server/store/mock"
	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterNodesV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	RegisterNodesV1(router, &zerolog.Logger{}, nil, store)

	assert.Equal(t, 10, len(router.Routes()))
}

func TestGetNodesHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/nodes", GetNodesHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListNodes(gomock.Any()).Return([]*structs.Node{
			{
				ID:           "1",
				Group:        &structs.Group{Name: "group1"},
				Name:         "node1",
				Organisation: "org1",
				Status:       structs.NodeStatusOnline,
				CPU:          structs.Resource{Capacity: 1, OvercommitRatio: 1.0},
				RAM:          structs.Resource{Capacity: 1024, OvercommitRatio: 1.0},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			{
				ID:           "2",
				Group:        &structs.Group{Name: "group1"},
				Name:         "node2",
				Organisation: "org1",
				Status:       structs.NodeStatusOnline,
				CPU:          structs.Resource{Capacity: 1, OvercommitRatio: 1.0},
				RAM:          structs.Resource{Capacity: 1024, OvercommitRatio: 1.0},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/nodes", nil)
		router.ServeHTTP(rec, req)
	})

	t.Run("success with organisation query", func(t *testing.T) {
		store.EXPECT().ListNodes(gomock.Any()).Return([]*structs.Node{
			{
				ID:           "1",
				Group:        &structs.Group{Name: "group1"},
				Name:         "node1",
				Organisation: "org1",
				Status:       structs.NodeStatusOnline,
				CPU:          structs.Resource{Capacity: 1, OvercommitRatio: 1.0},
				RAM:          structs.Resource{Capacity: 1024, OvercommitRatio: 1.0},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			{
				ID:           "2",
				Group:        &structs.Group{Name: "group1"},
				Name:         "node2",
				Organisation: "org2",
				Status:       structs.NodeStatusOnline,
				CPU:          structs.Resource{Capacity: 1, OvercommitRatio: 1.0},
				RAM:          structs.Resource{Capacity: 1024, OvercommitRatio: 1.0},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/nodes?organisation=org1", nil)
		router.ServeHTTP(rec, req)

		type response struct {
			Nodes []*structs.Node `json:"nodes"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(resp.Nodes))
		assert.Equal(t, "org1", resp.Nodes[0].Organisation)
	})

	t.Run("success with group query", func(t *testing.T) {
		store.EXPECT().ListNodes(gomock.Any()).Return([]*structs.Node{
			{
				ID:           "1",
				Group:        &structs.Group{Name: "group1"},
				Name:         "node1",
				Organisation: "org1",
				Status:       structs.NodeStatusOnline,
				CPU:          structs.Resource{Capacity: 1, OvercommitRatio: 1.0},
				RAM:          structs.Resource{Capacity: 1024, OvercommitRatio: 1.0},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			{
				ID:           "2",
				Group:        &structs.Group{Name: "group2"},
				Name:         "node2",
				Organisation: "org2",
				Status:       structs.NodeStatusOnline,
				CPU:          structs.Resource{Capacity: 1, OvercommitRatio: 1.0},
				RAM:          structs.Resource{Capacity: 1024, OvercommitRatio: 1.0},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/nodes?group=group2", nil)
		router.ServeHTTP(rec, req)

		type response struct {
			Nodes []*structs.Node `json:"nodes"`
		}

		var resp response
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(resp.Nodes))
		assert.Equal(t, "group2", resp.Nodes[0].Group.Name)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().ListNodes(gomock.Any()).Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/nodes", nil)
		router.ServeHTTP(rec, req)
	})
}

func TestGetNodeHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/nodes/:id", GetNodeHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetNode(gomock.Any(), "1").Return(&structs.Node{
			ID:           "1",
			Group:        &structs.Group{Name: "group1"},
			Name:         "node1",
			Organisation: "org1",
			Status:       structs.NodeStatusOnline,
			CPU:          structs.Resource{Capacity: 1, OvercommitRatio: 1.0},
			RAM:          structs.Resource{Capacity: 1024, OvercommitRatio: 1.0},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/nodes/1", nil)
		router.ServeHTTP(rec, req)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().GetNode(gomock.Any(), "1").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/nodes/1", nil)
		router.ServeHTTP(rec, req)
	})
}
