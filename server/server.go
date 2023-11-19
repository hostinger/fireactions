package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/server/github"
	"github.com/hostinger/fireactions/server/scheduler"
	"github.com/hostinger/fireactions/server/store"
	"github.com/hostinger/fireactions/server/store/bbolt"
	"github.com/hostinger/fireactions/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// Server struct.
type Server struct {
	store        store.Store
	scheduler    *scheduler.Scheduler
	server       *http.Server
	config       *Config
	github       *github.Client
	shutdownOnce sync.Once
	shutdownMu   sync.Mutex
	shutdownCh   chan struct{}
	logger       *zerolog.Logger

	up       prometheus.Gauge
	registry *prometheus.Registry
}

// New creates a new Server.
func New(config *Config) (*Server, error) {
	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	logger := zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().CallerWithSkipFrameCount(2).Logger()

	s := &Server{
		config:       config,
		shutdownOnce: sync.Once{},
		shutdownCh:   make(chan struct{}),
		shutdownMu:   sync.Mutex{},
		registry:     prometheus.NewRegistry(),
		logger:       &logger,
	}

	mux := gin.New()
	mux.Use(requestid.New(requestid.WithCustomHeaderStrKey("X-Request-ID")))
	mux.Use(gin.Recovery())

	store, err := bbolt.New(filepath.Join(config.DataDir, "fireactions.db"))
	if err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}
	s.store = store

	scheduler, err := scheduler.New(logger, store)
	if err != nil {
		return nil, fmt.Errorf("error creating scheduler: %w", err)
	}
	s.scheduler = scheduler

	github, err := github.NewClient(&github.ClientConfig{AppID: config.GitHubConfig.AppID, AppPrivateKey: config.GitHubConfig.AppPrivateKey})
	if err != nil {
		return nil, fmt.Errorf("creating GitHub client: %w", err)
	}
	s.github = github

	mux.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))
	mux.GET("/version", s.handleGetVersion)
	mux.GET("/healthz", s.handleGetHealthz)

	mux.POST("/webhook", s.handleGitHubWebhook())
	v1 := mux.Group("/api/v1")
	{
		v1.GET("/healthz", s.handleGetHealthz)
		v1.GET("/runners", s.handleGetRunners)
		v1.GET("/runners/:id", s.handleGetRunner)
		v1.GET("/nodes", s.handleGetNodes)
		v1.POST("/nodes", s.handleRegisterNode)
		v1.GET("/nodes/:id", s.handleGetNode)
		v1.POST("/runners", s.handleCreateRunner)
		v1.PATCH("/runners/:id/status", s.handleSetRunnerStatus)
		v1.GET("/runners/:id/registration-token", s.handleGetRunnerRegistrationToken)
		v1.GET("/runners/:id/remove-token", s.handleGetRunnerRemoveToken)
		v1.DELETE("/runners/:id", s.handleDeleteRunner)
		v1.GET("/nodes/:id/runners", s.handleGetNodeRunners)
		v1.DELETE("/nodes/:id", s.handleDeregisterNode)
		v1.POST("/nodes/:id/cordon", s.handleCordonNode)
		v1.POST("/nodes/:id/uncordon", s.handleUncordonNode)
	}

	s.server = &http.Server{
		Addr:         config.HTTP.ListenAddress,
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.up = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "up",
		Help:      "Whether the server is up.",
		Namespace: "fireactions",
	})

	s.registry.MustRegister(s)

	return s, nil
}

// Shutdown attempts to gracefully shutdown the Server. It blocks until the Server is shutdown or the context is
// cancelled.
func (s *Server) Shutdown(ctx context.Context) {
	s.shutdownOnce.Do(func() {
		s.logger.Info().Msg("Stopping server")

		err := s.server.Shutdown(ctx)
		if err != nil {
			s.logger.Error().Err(err).Msg("error shutting down server")
		}

		s.scheduler.Shutdown()
		s.store.Close()

		close(s.shutdownCh)
	})
}

// Start starts the Server. It blocks until Shutdown() is called.
func (s *Server) Start() error {
	s.logger.Info().Str("version", version.Version).Str("date", version.Date).Str("commit", version.Commit).Msgf("Starting server on %s", s.config.HTTP.ListenAddress)

	var err error

	err = s.scheduler.Start()
	if err != nil {
		return fmt.Errorf("scheduler: %w", err)
	}

	s.up.Set(1)

	err = s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Collect implements the prometheus.Collector interface.
func (s *Server) Collect(ch chan<- prometheus.Metric) {
	s.up.Collect(ch)
}

// Describe implements the prometheus.Collector interface.
func (s *Server) Describe(ch chan<- *prometheus.Desc) {
	s.up.Describe(ch)
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}
