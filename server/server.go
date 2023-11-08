package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/server/config"
	"github.com/hostinger/fireactions/server/github"
	"github.com/hostinger/fireactions/server/handlers"
	"github.com/hostinger/fireactions/server/httperr"
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
	TLSConfig    *tls.Config
	store        store.Store
	scheduler    *scheduler.Scheduler
	server       *http.Server
	config       *config.Config
	shutdownOnce sync.Once
	shutdownMu   sync.Mutex
	shutdownCh   chan struct{}
	logger       *zerolog.Logger

	up       prometheus.Gauge
	registry *prometheus.Registry
}

// New creates a new Server.
func New(config *config.Config) (*Server, error) {
	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	logger := zerolog.
		New(os.Stdout).Level(logLevel).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	s := &Server{
		TLSConfig:    &tls.Config{},
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
	mux.Use(httperr.HandlerFunc(
		httperr.Map(store.ErrNotFound).To(http.StatusNotFound, "Resource doesn't exist."),
	))

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

	tokenGetter, err := github.NewClient(&github.ClientConfig{
		AppID:         config.GitHubConfig.AppID,
		AppPrivateKey: config.GitHubConfig.AppPrivateKey,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub client: %w", err)
	}

	v1 := mux.Group("/api/v1")
	{
		handlers.RegisterMiscHandlers(v1)
		handlers.RegisterRunnersHandlers(
			&logger, v1, scheduler, store, tokenGetter, config)
		handlers.RegisterNodesHandlers(
			&logger, v1, scheduler, store)
	}

	mux.POST("/webhook", handlers.GitHubWebhookHandlerFunc(&logger, store, scheduler, config.GitHubConfig))
	mux.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))

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
		s.logger.Info().Msg("stopping server")

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
	s.logger.Info().Str("version", version.Version).Str("date", version.Date).Str("commit", version.Commit).Msgf("starting server on %s", s.config.HTTP.ListenAddress)

	var err error

	err = s.scheduler.Start()
	if err != nil {
		return fmt.Errorf("error starting scheduler: %w", err)
	}

	s.up.Set(1)

	err = s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting server: %w", err)
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
