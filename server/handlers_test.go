package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func TestGetHealthzHandler(t *testing.T) {
	router := gin.New()
	router.GET("/healthz", getHealthzHandler())

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}
}

func TestGetVersionHandler(t *testing.T) {
	router := gin.New()
	router.GET("/version", getVersionHandler())

	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}
}

func TestListPoolsHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("EmptyList", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().ListPools(gomock.Any()).Return([]*Pool{}, nil)

		router := gin.New()
		router.GET("/api/v1/pools", listPoolsHandler(m))

		req, err := http.NewRequest("GET", "/api/v1/pools", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
		}

		expectedBody := `{"pools":[]}`
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().ListPools(gomock.Any()).Return(nil, errors.New("error"))

		router := gin.New()
		router.GET("/api/v1/pools", listPoolsHandler(m))

		req, err := http.NewRequest("GET", "/api/v1/pools", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, but got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestGetPoolHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("Success", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().GetPool(gomock.Any(), "test").Return(&Pool{
			machines: make(map[string]*firecracker.Machine),
			config: &PoolConfig{
				Name:       "test",
				MaxRunners: 0,
				MinRunners: 0,
			},
			isActive: false,
		}, nil)

		router := gin.New()
		router.GET("/api/v1/pools/:id", getPoolHandler(m))

		req, err := http.NewRequest("GET", "/api/v1/pools/test", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
		}

		expectedBody := `{"pool":{"name":"test","max_runners":0,"min_runners":0,"cur_runners":0,"status":{"state":"Paused","message":"Pool is paused"}}}`
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().GetPool(gomock.Any(), "test").Return(nil, errors.New("error"))

		router := gin.New()
		router.GET("/api/v1/pools/:id", getPoolHandler(m))

		req, err := http.NewRequest("GET", "/api/v1/pools/test", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func TestScalePoolHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("Success", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().ScalePool(gomock.Any(), "test", 1).Return(nil)

		router := gin.New()
		router.POST("/api/v1/pools/:id/scale", scalePoolHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/pools/test/scale", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
		}

		expectedBody := `{"message":"Pool scaled successfully"}`
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().ScalePool(gomock.Any(), "test", 1).Return(errors.New("error"))

		router := gin.New()
		router.POST("/api/v1/pools/:id/scale", scalePoolHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/pools/test/scale", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, but got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestPausePoolHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("Success", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().PausePool(gomock.Any(), "test").Return(nil)

		router := gin.New()
		router.POST("/api/v1/pools/:id/pause", pausePoolHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/pools/test/pause", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
		}

		expectedBody := `{"message":"Pool paused successfully"}`
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().PausePool(gomock.Any(), "test").Return(errors.New("error"))

		router := gin.New()
		router.POST("/api/v1/pools/:id/pause", pausePoolHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/pools/test/pause", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, but got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestResumePoolHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("Success", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().ResumePool(gomock.Any(), "test").Return(nil)

		router := gin.New()
		router.POST("/api/v1/pools/:id/resume", resumePoolHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/pools/test/resume", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
		}

		expectedBody := `{"message":"Pool resumed successfully"}`
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().ResumePool(gomock.Any(), "test").Return(errors.New("error"))

		router := gin.New()
		router.POST("/api/v1/pools/:id/resume", resumePoolHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/pools/test/resume", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, but got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestRestartHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("Success", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().Restart(gomock.Any()).Return(nil)

		router := gin.New()
		router.POST("/api/v1/restart", restartHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/restart", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
		}

		expectedBody := `{"message":"Pools restarted successfully"}`
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		m := newMockPoolManager(mockCtrl)
		m.EXPECT().Restart(gomock.Any()).Return(errors.New("error"))

		router := gin.New()
		router.POST("/api/v1/restart", restartHandler(m))

		req, err := http.NewRequest("POST", "/api/v1/restart", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, but got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func init() {
	gin.SetMode(gin.TestMode)
}
