package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/helper/github"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// Server represents the Fireactions server.
type Server struct {
	config        *Config
	pools         map[string]*Pool
	server        *http.Server
	metricsServer *http.Server
	github        *github.Client
	l             *sync.Mutex
	logger        *zerolog.Logger
}

// Opt is a functional option for Server.
type Opt func(s *Server)

// WithLogger sets the logger for the Server.
func WithLogger(logger *zerolog.Logger) Opt {
	f := func(s *Server) {
		s.logger = logger
	}

	return f
}

// New creates a new Server.
func New(config *Config, opts ...Opt) (*Server, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	github, err := github.NewClient(config.GitHub.AppID, config.GitHub.AppPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("creating github client: %w", err)
	}

	gin.SetMode(gin.ReleaseMode)
	handler := gin.New()
	handler.Use(requestid.New(requestid.WithCustomHeaderStrKey("X-Request-ID")))
	handler.Use(gin.Recovery())

	server := &http.Server{
		Addr:         config.BindAddress,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger := zerolog.Nop()
	s := &Server{
		config: config,
		server: server,
		pools:  make(map[string]*Pool),
		github: github,
		l:      &sync.Mutex{},
		logger: &logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	if config.Metrics.Enabled {
		metricsHandler := http.NewServeMux()
		metricsHandler.Handle("/metrics", promhttp.Handler())
		metricsServer := &http.Server{
			Addr:         config.Metrics.Address,
			Handler:      metricsHandler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		s.metricsServer = metricsServer
	}

	handler.POST("/webhook/github", webhookGitHubHandler(s, config.GitHub.WebhookSecret))
	handler.GET("/healthz", getHealthzHandler())
	handler.GET("/version", getVersionHandler())

	if config.Debug {
		pprof.Register(handler)
	}

	api := handler.Group("/api")
	if config.BasicAuthEnabled {
		api.Use(gin.BasicAuth(gin.Accounts(config.BasicAuthUsers)))
	}

	v1 := api.Group("/v1")
	{
		v1.GET("/pools", listPoolsHandler(s))
		v1.POST("/pools/:id/scale", scalePoolHandler(s))
		v1.GET("/pools/:id", getPoolHandler(s))
		v1.POST("/pools/:id/resume", resumePoolHandler(s))
		v1.POST("/pools/:id/pause", pausePoolHandler(s))
		v1.POST("/restart", restartHandler(s))
	}

	return s, nil
}

// Run starts the server and blocks until the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info().Str("version", fireactions.Version).Str("date", fireactions.Date).Str("commit", fireactions.Commit).Msgf("Starting server on %s", s.config.BindAddress)
	if s.config.Debug {
		s.logger.Warn().Msg("Debug mode enabled")
	}

	listener, err := net.Listen("tcp", s.config.BindAddress)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer listener.Close()

	for _, poolConfig := range s.config.Pools {
		pool, err := NewPool(s.logger, poolConfig, s.github)
		if err != nil {
			return fmt.Errorf("creating pool: %w", err)
		}

		s.pools[poolConfig.Name] = pool
		go pool.Start()
		s.logger.Info().Msgf("Pool %s started", poolConfig.Name)
	}

	errGroup := &errgroup.Group{}
	errGroup.Go(func() error { return s.server.Serve(listener) })
	if s.metricsServer != nil {
		metricsListener, err := net.Listen("tcp", s.config.Metrics.Address)
		if err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}

		errGroup.Go(func() error { return s.metricsServer.Serve(metricsListener) })
	}

	go func() {
		<-ctx.Done()
		fmt.Println()

		s.logger.Info().Msg("Shutting down server")

		wg := sync.WaitGroup{}
		for _, pool := range s.pools {
			wg.Add(1)
			go func(pool *Pool) {
				pool.Stop()
				wg.Done()
			}(pool)
		}

		wg.Wait()

		cancelCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if s.config.Metrics.Enabled {
			_ = s.metricsServer.Shutdown(cancelCtx)
		}

		if err := s.server.Shutdown(cancelCtx); err != nil {
			s.logger.Error().Err(err).Msg("Failed to shutdown server")
		}
	}()

	metricUp.Set(1)

	err = errGroup.Wait()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	s.logger.Info().Msg("Server stopped")
	return nil
}

// GetPool returns the pool with the given ID.
func (s *Server) GetPool(ctx context.Context, id string) (*Pool, error) {
	s.l.Lock()
	defer s.l.Unlock()

	pool, ok := s.pools[id]
	if !ok {
		return nil, fireactions.ErrPoolNotFound
	}

	return pool, nil
}

// ListPools returns a list of all pools.
func (s *Server) ListPools(ctx context.Context) ([]*Pool, error) {
	s.l.Lock()
	defer s.l.Unlock()

	pools := make([]*Pool, 0, len(s.pools))
	for _, pool := range s.pools {
		pools = append(pools, pool)
	}

	return pools, nil
}

// ScalePool scales the pool with the given ID to the desired size.
func (s *Server) ScalePool(ctx context.Context, id string, replicas int) error {
	metricPoolScaleRequests.WithLabelValues(id).Inc()

	pool, err := s.GetPool(ctx, id)
	if err != nil {
		return err
	}

	return pool.Scale(ctx, replicas)
}

// PausePool pauses the pool with the given ID.
func (s *Server) PausePool(ctx context.Context, id string) error {
	pool, err := s.GetPool(ctx, id)
	if err != nil {
		return err
	}

	pool.Pause()
	metricPoolStatus.WithLabelValues(id).Set(0)
	return nil
}

// ResumePool resumes the pool with the given ID.
func (s *Server) ResumePool(ctx context.Context, id string) error {
	pool, err := s.GetPool(ctx, id)
	if err != nil {
		return err
	}

	pool.Resume()
	metricPoolStatus.WithLabelValues(id).Set(1)
	return nil
}

func (s *Server) Restart(ctx context.Context) error {
	s.l.Lock()
	defer s.l.Unlock()

	s.logger.Info().Msgf("Restarting server configuration")
	err := s.config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	for _, poolConfig := range s.config.Pools {
		pool, ok := s.pools[poolConfig.Name]
		if ok {
			pool.config = poolConfig
			s.logger.Info().Msgf("Pool %s reloaded", poolConfig.Name)
			continue
		}

		pool, err = NewPool(s.logger, poolConfig, s.github)
		if err != nil {
			return fmt.Errorf("creating pool: %w", err)
		}

		s.pools[poolConfig.Name] = pool
		go pool.Start()
		s.logger.Info().Msgf("Pool %s started", poolConfig.Name)
	}

	return nil
}
