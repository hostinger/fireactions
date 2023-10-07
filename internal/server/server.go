package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/internal/server/ghclient"
	"github.com/hostinger/fireactions/internal/server/handler"
	"github.com/hostinger/fireactions/internal/server/scheduler"
	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/server/structs"
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
	shutdownOnce sync.Once
	shutdownMu   sync.Mutex
	shutdownCh   chan struct{}
	log          *zerolog.Logger
	cfg          *Config

	up       prometheus.Gauge
	registry *prometheus.Registry
}

// New creates a new Server.
func New(log *zerolog.Logger, cfg *Config, store store.Store) (*Server, error) {
	s := &Server{
		TLSConfig:    &tls.Config{},
		cfg:          cfg,
		shutdownOnce: sync.Once{},
		shutdownCh:   make(chan struct{}),
		shutdownMu:   sync.Mutex{},
		store:        store,
		scheduler:    scheduler.New(log, cfg.Scheduler, store),
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "up",
			Namespace: "fireactions",
			Subsystem: "server",
			Help:      "Whether the server is up",
		}),
		registry: prometheus.NewRegistry(),
	}

	logger := log.With().Str("component", "server").Logger()
	s.log = &logger

	ghClient, err := ghclient.New(&ghclient.Config{
		AppID:         cfg.GitHub.AppID,
		AppPrivateKey: cfg.GitHub.AppPrivateKey,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub client: %w", err)
	}

	mux := gin.New()
	mux.Use(requestid.New(requestid.WithCustomHeaderStrKey("X-Request-ID")))
	mux.Use(gin.Recovery())

	v1 := mux.Group("/api/v1")
	{
		handler.RegisterJobsV1(v1, log, store)
		handler.RegisterFlavorsV1(v1, log, store)
		handler.RegisterGitHubV1(v1, log, ghClient)
		handler.RegisterGroupsV1(v1, log, store)
		handler.RegisterRunnersV1(v1, log, store)
		handler.RegisterNodesV1(v1, log, s.scheduler, store)
	}

	mux.GET("/healthz", handler.GetHealthzHandlerFunc())
	mux.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))
	mux.POST("/webhook", handler.GetGitHubWebhookHandlerFuncV1(
		log, cfg.GitHub.WebhookSecret, cfg.GitHub.JobLabelPrefix, cfg.DefaultFlavor, cfg.DefaultGroup, s.scheduler, store))

	s.server = &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// Shutdown attempts to gracefully shutdown the Server. It blocks until the Server is shutdown or the context is
// cancelled.
func (s *Server) Shutdown(ctx context.Context) {
	s.shutdownOnce.Do(func() {
		err := s.server.Shutdown(ctx)
		if err != nil {
			s.log.Error().Err(err).Msg("error shutting down server")
		}

		s.scheduler.Shutdown()
		s.store.Close()

		close(s.shutdownCh)
	})
}

// Start starts the Server. It blocks until Shutdown() is called.
func (s *Server) Start() error {
	s.log.Info().Msgf("starting server on %s", s.server.Addr)

	var err error

	err = s.init()
	if err != nil {
		return fmt.Errorf("error initializing server: %w", err)
	}

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

func (s *Server) init() error {
	s.registry.MustRegister(s.store)
	s.registry.MustRegister(s)

	var err error

	err = s.initPreconfiguredFlavors()
	if err != nil {
		return fmt.Errorf("error initializing preconfigured flavors: %w", err)
	}

	err = s.initPreconfiguredGroups()
	if err != nil {
		return fmt.Errorf("error initializing preconfigured groups: %w", err)
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

func (s *Server) initPreconfiguredFlavors() error {
	s.log.Info().Msg("creating preconfigured Flavor(s)")

	for _, cfg := range s.cfg.Flavors {
		err := s.store.SaveFlavor(context.Background(), &structs.Flavor{
			Name:         cfg.Name,
			Enabled:      *cfg.Enabled,
			VCPUs:        cfg.CPU,
			MemorySizeMB: cfg.Mem,
			DiskSizeGB:   cfg.Disk,
			Image:        cfg.Image,
		})
		if err != nil {
			return fmt.Errorf("error creating preconfigured flavor (%s): %w", cfg.Name, err)
		}

		s.log.Info().Msgf("created preconfigured flavor (%s)", cfg.Name)
	}

	return nil
}

func (s *Server) initPreconfiguredGroups() error {
	s.log.Info().Msg("creating preconfigured Group(s)")

	for _, cfg := range s.cfg.Groups {
		err := s.store.SaveGroup(context.Background(), &structs.Group{
			Name:    cfg.Name,
			Enabled: *cfg.Enabled,
		})
		if err != nil {
			return fmt.Errorf("error creating preconfigured group (%s): %w", cfg.Name, err)
		}

		s.log.Info().Msgf("created preconfigured group (%s)", cfg.Name)
	}

	return nil
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}
