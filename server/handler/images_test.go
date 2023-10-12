package handler

import (
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

func TestRegisterImagesV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	RegisterImagesV1(router, &zerolog.Logger{}, store)

	assert.Equal(t, 3, len(router.Routes()))
}

func TestGetImagesHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/images", GetImagesHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().ListImages(gomock.Any()).Return([]*structs.Image{
			{
				ID:   "48233fc0-8c16-491b-8666-970ba3ce771e",
				Name: "image-1",
				URL:  "https://example.com/image-1",
			},
			{
				ID:   "48233fc0-8c16-491b-8666-970ba3ce771f",
				Name: "image-2",
				URL:  "https://example.com/image-2",
			},
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/images", nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		type Root struct {
			Images []*structs.Image `json:"images"`
		}

		var root Root
		err := json.Unmarshal(rec.Body.Bytes(), &root)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(root.Images))
		assert.Equal(t, "48233fc0-8c16-491b-8666-970ba3ce771e", root.Images[0].ID)
		assert.Equal(t, "image-1", root.Images[0].Name)
		assert.Equal(t, "https://example.com/image-1", root.Images[0].URL)
		assert.Equal(t, "48233fc0-8c16-491b-8666-970ba3ce771f", root.Images[1].ID)
		assert.Equal(t, "image-2", root.Images[1].Name)
		assert.Equal(t, "https://example.com/image-2", root.Images[1].URL)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().ListImages(gomock.Any()).Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/images", nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestGetImageHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.GET("/images/:id", GetImageHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetImage(gomock.Any(), "48233fc0-8c16-491b-8666-970ba3ce771e").Return(&structs.Image{
			ID:   "48233fc0-8c16-491b-8666-970ba3ce771e",
			Name: "image-1",
			URL:  "https://example.com/image-1",
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/images/48233fc0-8c16-491b-8666-970ba3ce771e", nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		var image structs.Image
		err := json.Unmarshal(rec.Body.Bytes(), &image)

		assert.NoError(t, err)
		assert.Equal(t, "48233fc0-8c16-491b-8666-970ba3ce771e", image.ID)
		assert.Equal(t, "image-1", image.Name)
		assert.Equal(t, "https://example.com/image-1", image.URL)
	})

	t.Run("error", func(t *testing.T) {
		store.EXPECT().GetImage(gomock.Any(), "48233fc0-8c16-491b-8666-970ba3ce771e").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/images/48233fc0-8c16-491b-8666-970ba3ce771e", nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestDeleteImageHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStore(ctrl)

	router := gin.New()
	router.DELETE("/images/:id", DeleteImageHandlerFuncV1(&zerolog.Logger{}, store))

	t.Run("success", func(t *testing.T) {
		store.EXPECT().GetImageByID(gomock.Any(), "48233fc0-8c16-491b-8666-970ba3ce771e").Return(&structs.Image{
			ID:   "48233fc0-8c16-491b-8666-970ba3ce771e",
			Name: "image-1",
			URL:  "https://example.com/image-1",
		}, nil)

		store.EXPECT().DeleteImage(gomock.Any(), "48233fc0-8c16-491b-8666-970ba3ce771e").Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/images/48233fc0-8c16-491b-8666-970ba3ce771e", nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, 204, rec.Code)
	})

	t.Run("error on GetImageByID", func(t *testing.T) {
		store.EXPECT().GetImageByID(gomock.Any(), "48233fc0-8c16-491b-8666-970ba3ce771e").Return(nil, errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/images/48233fc0-8c16-491b-8666-970ba3ce771e", nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})

	t.Run("error on DeleteImage", func(t *testing.T) {
		store.EXPECT().GetImageByID(gomock.Any(), "48233fc0-8c16-491b-8666-970ba3ce771e").Return(&structs.Image{
			ID:   "48233fc0-8c16-491b-8666-970ba3ce771e",
			Name: "image-1",
			URL:  "https://example.com/image-1",
		}, nil)

		store.EXPECT().DeleteImage(gomock.Any(), "48233fc0-8c16-491b-8666-970ba3ce771e").Return(errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/images/48233fc0-8c16-491b-8666-970ba3ce771e", nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}
