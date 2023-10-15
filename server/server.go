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
	"github.com/hostinger/fireactions/server/ghclient"
	"github.com/hostinger/fireactions/server/handler"
	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/scheduler"
	"github.com/hostinger/fireactions/server/store"
	"github.com/hostinger/fireactions/server/store/bbolt"
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
	config       *Config
	shutdownOnce sync.Once
	shutdownMu   sync.Mutex
	shutdownCh   chan struct{}
	logger       *zerolog.Logger

	up       prometheus.Gauge
	registry *prometheus.Registry
}

// New creates a new Server.
func New(cfg *Config) (*Server, error) {
	store, err := bbolt.New(filepath.Join(cfg.DataDir, "fireactions.db"))
	if err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	logger := zerolog.
		New(os.Stdout).Level(logLevel).With().Timestamp().Str("component", "server").Logger()

	scheduler, err := scheduler.New(logger, cfg.Scheduler, store)
	if err != nil {
		return nil, fmt.Errorf("error creating scheduler: %w", err)
	}

	s := &Server{
		TLSConfig:    &tls.Config{},
		config:       cfg,
		scheduler:    scheduler,
		shutdownOnce: sync.Once{},
		shutdownCh:   make(chan struct{}),
		shutdownMu:   sync.Mutex{},
		store:        store,
		registry:     prometheus.NewRegistry(),
		logger:       &logger,
	}

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
		handler.RegisterJobsV1(v1, &logger, store)
		handler.RegisterFlavorsV1(v1, &logger, store)
		handler.RegisterGitHubV1(v1, &logger, ghClient)
		handler.RegisterGroupsV1(v1, &logger, store)
		handler.RegisterRunnersV1(v1, &logger, store)
		handler.RegisterNodesV1(v1, &logger, s.scheduler, store)
		handler.RegisterImagesV1(v1, &logger, store)
	}

	mux.GET("/healthz", handler.GetHealthzHandlerFunc())
	mux.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))
	mux.POST("/webhook", handler.GetGitHubWebhookHandlerFuncV1(
		&logger, cfg.GitHub.WebhookSecret, cfg.GitHub.JobLabelPrefix, s.scheduler, store))

	s.server = &http.Server{
		Addr:         cfg.ListenAddr,
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

	return s, nil
}

// Shutdown attempts to gracefully shutdown the Server. It blocks until the Server is shutdown or the context is
// cancelled.
func (s *Server) Shutdown(ctx context.Context) {
	s.shutdownOnce.Do(func() {
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
	s.logger.Info().Msgf("starting server on %s", s.server.Addr)

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

	err = s.initPreconfiguredImages()
	if err != nil {
		return fmt.Errorf("error initializing preconfigured images: %w", err)
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
	s.logger.Info().Msg("creating preconfigured Flavor(s)")

	for _, cfg := range s.config.Flavors {
		err := s.store.SaveFlavor(context.Background(), &models.Flavor{
			Name:         cfg.Name,
			Enabled:      cfg.Enabled,
			VCPUs:        cfg.CPU,
			MemorySizeMB: cfg.Mem,
			DiskSizeGB:   cfg.Disk,
			Image:        cfg.Image,
		})
		if err != nil {
			return fmt.Errorf("error creating preconfigured flavor (%s): %w", cfg.Name, err)
		}

		s.logger.Info().Msgf("created preconfigured flavor (%s)", cfg.Name)
	}

	defaultFlavor, err := s.store.GetFlavor(context.Background(), s.config.DefaultFlavor)
	if err != nil {
		return fmt.Errorf("error fetching default flavor: %w", err)
	}

	err = s.store.SetDefaultFlavor(context.Background(), defaultFlavor.Name)
	if err != nil {
		return fmt.Errorf("error setting default flavor: %w", err)
	}

	s.logger.Info().Msgf("set default flavor to %s", defaultFlavor.Name)
	return nil
}

func (s *Server) initPreconfiguredGroups() error {
	s.logger.Info().Msg("creating preconfigured Group(s)")

	for _, cfg := range s.config.Groups {
		err := s.store.SaveGroup(context.Background(), &models.Group{
			Name:    cfg.Name,
			Enabled: cfg.Enabled,
		})
		if err != nil {
			return fmt.Errorf("error creating preconfigured group (%s): %w", cfg.Name, err)
		}

		s.logger.Info().Msgf("created preconfigured group (%s)", cfg.Name)
	}

	defaultGroup, err := s.store.GetGroup(context.Background(), s.config.DefaultGroup)
	if err != nil {
		return fmt.Errorf("error fetching default group: %w", err)
	}

	err = s.store.SetDefaultGroup(context.Background(), defaultGroup.Name)
	if err != nil {
		return fmt.Errorf("error setting default group: %w", err)
	}

	s.logger.Info().Msgf("set default group to %s", defaultGroup.Name)
	return nil
}

func (s *Server) initPreconfiguredImages() error {
	s.logger.Info().Msg("creating preconfigured Image(s)")

	for _, cfg := range s.config.Images {
		err := s.store.SaveImage(context.Background(), &models.Image{
			ID:   cfg.ID,
			Name: cfg.Name,
			URL:  cfg.URL,
		})
		if err != nil {
			return fmt.Errorf("error creating preconfigured image (%s): %w", cfg.Name, err)
		}

		s.logger.Info().Msgf("created preconfigured image (%s)", cfg.Name)
	}

	return nil
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}
