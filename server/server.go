package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/server/github"
	"github.com/hostinger/fireactions/server/metric"
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
	store     store.Store
	scheduler *scheduler.Scheduler
	server    *http.Server
	config    *Config
	github    *github.Client
	logger    *zerolog.Logger
}

// New creates a new Server.
func New(config *Config) (*Server, error) {
	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	logger := zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().CallerWithSkipFrameCount(2).Logger()

	store, err := bbolt.New(filepath.Join(config.DataDir, "fireactions.db"))
	if err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}

	scheduler, err := scheduler.New(logger, store)
	if err != nil {
		return nil, fmt.Errorf("error creating scheduler: %w", err)
	}

	github, err := github.NewClient(&github.ClientConfig{
		AppID: config.GitHub.AppID, AppPrivateKey: config.GitHub.AppPrivateKey,
	})
	if err != nil {
		return nil, fmt.Errorf("creating GitHub client: %w", err)
	}

	mux := gin.New()
	mux.Use(requestid.New(requestid.WithCustomHeaderStrKey("X-Request-ID")))
	mux.Use(gin.Recovery())

	server := &http.Server{
		Addr:         config.HTTP.ListenAddress,
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(metric.NewPrometheusCollector(store, metric.WithLogger(&logger)))

	s := &Server{
		config:    config,
		github:    github,
		scheduler: scheduler,
		store:     store,
		server:    server,
		logger:    &logger,
	}

	mux.POST("/webhook", s.handleWebhook())
	mux.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	mux.GET("/version", s.handleGetVersion)
	mux.GET("/healthz", s.handleGetHealthz)

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
		v1.DELETE("/runners/:id", s.handleDeleteRunner)
		v1.GET("/nodes/:id/runners", s.handleGetNodeRunners)
		v1.DELETE("/nodes/:id", s.handleDeregisterNode)
		v1.POST("/nodes/:id/cordon", s.handleCordonNode)
		v1.POST("/nodes/:id/uncordon", s.handleUncordonNode)
	}

	return s, nil
}

// Run starts the server. It blocks until the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info().Str("version", version.Version).Str("date", version.Date).Str("commit", version.Commit).Msgf("Starting server on %s", s.config.HTTP.ListenAddress)

	var err error
	err = s.scheduler.Start()
	if err != nil {
		return fmt.Errorf("scheduler: %w", err)
	}
	defer s.scheduler.Shutdown()

	go func() {
		<-ctx.Done()
		fmt.Println()

		s.logger.Info().Msg("Stopping server")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := s.server.Shutdown(ctx)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to shutdown server")
		}

		s.logger.Info().Msg("Server stopped")
	}()

	err = s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}
