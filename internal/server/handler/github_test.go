package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/internal/server/handler/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetGitHubRegistrationTokenHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tg := mock.NewMockGitHubTokenGetter(ctrl)

	router := gin.New()
	RegisterGitHubV1(router, &zerolog.Logger{}, tg)

	t.Run("success", func(t *testing.T) {
		tg.EXPECT().GetRegistrationToken(gomock.Any(), "org").Return("token", nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/github/org/registration-token", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		assert.JSONEq(t, `{"token":"token"}`, rec.Body.String())
	})

	t.Run("error", func(t *testing.T) {
		tg.EXPECT().GetRegistrationToken(gomock.Any(), "org").Return("", errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/github/org/registration-token", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}

func TestGetGitHubRemoveTokenHandlerFuncV1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tg := mock.NewMockGitHubTokenGetter(ctrl)

	router := gin.New()
	RegisterGitHubV1(router, &zerolog.Logger{}, tg)

	t.Run("success", func(t *testing.T) {
		tg.EXPECT().GetRemoveToken(gomock.Any(), "org").Return("token", nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/github/org/remove-token", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
		assert.JSONEq(t, `{"token":"token"}`, rec.Body.String())
	})

	t.Run("error", func(t *testing.T) {
		tg.EXPECT().GetRemoveToken(gomock.Any(), "org").Return("", errors.New("error"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/github/org/remove-token", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, 500, rec.Code)
		assert.JSONEq(t, `{"error":"error"}`, rec.Body.String())
	})
}
