package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/containerd/containerd"
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
	containerd    *containerd.Client
	imageManager  *imageManager
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

	containerdClient, err := containerd.New(config.Containerd.Address,
		containerd.WithTimeout(5*time.Second), containerd.WithDefaultNamespace(config.Containerd.Namespace))
	if err != nil {
		return nil, fmt.Errorf("containerd: creating client: %w", err)
	}

	s := &Server{
		config:     config,
		server:     server,
		pools:      make(map[string]*Pool),
		github:     github,
		containerd: containerdClient,
		l:          &sync.Mutex{},
		logger:     &logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.imageManager = newImageManager(s.logger, containerdClient)

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
		v1.POST("/reload", reloadHandler(s))
		v1.GET("/pools/:id/microvms", listMicroVMsHandler(s))
		v1.GET("/microvms", listMicroVMsHandler(s)) // List all microvms across all pools
		v1.GET("/microvms/:id", getMicroVMHandler(s))
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
	defer func() {
		_ = listener.Close()
	}()

	for _, poolConfig := range s.config.Pools {
		pool, err := NewPool(s.logger, poolConfig, s.github, s.imageManager, s.containerd)
		if err != nil {
			return fmt.Errorf("creating pool: %w", err)
		}

		s.pools[poolConfig.Name] = pool
		go pool.Run()
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

		// Stop pools sequentially to avoid lock contention and race conditions
		for name, pool := range s.pools {
			s.logger.Info().Msgf("Stopping pool %s", name)
			pool.Stop()
			s.logger.Info().Msgf("Pool %s stopped", name)
		}

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
	pool, err := s.GetPool(ctx, id)
	if err != nil {
		return err
	}

	metricPoolScaleRequests.WithLabelValues(id, pool.config.Runner.Organization).Inc()

	// Update the pool config with the new replicas value
	// The Run() loop will handle the actual scaling
	pool.SetReplicas(replicas)

	return nil
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

func (s *Server) Reload(ctx context.Context) error {
	s.l.Lock()
	defer s.l.Unlock()

	s.logger.Info().Msgf("Reloading server configuration")
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

		pool, err = NewPool(s.logger, poolConfig, s.github, s.imageManager, s.containerd)
		if err != nil {
			return fmt.Errorf("creating pool: %w", err)
		}

		s.pools[poolConfig.Name] = pool
		go pool.Run()
		s.logger.Info().Msgf("Pool %s started", poolConfig.Name)
	}

	return nil
}

// ListMicroVMs returns a list of MicroVMs for the given poolName.
func (s *Server) ListMicroVMs(ctx context.Context, poolName string) ([]*MicroVM, error) {
	if poolName == "" {
		s.l.Lock()
		pools := make([]*Pool, 0, len(s.pools))
		for _, pool := range s.pools {
			pools = append(pools, pool)
		}
		s.l.Unlock()

		var allVMs []*MicroVM
		for _, pool := range pools {
			vms, err := pool.ListMicroVMs(ctx, pool.config.Name)
			if err != nil {
				// Log error but continue with other pools
				s.logger.Warn().Err(err).Str("pool", pool.config.Name).Msg("Failed to list microvms for pool")
				continue
			}

			allVMs = append(allVMs, vms...)
		}

		return allVMs, nil
	}

	pool, err := s.GetPool(ctx, poolName)
	if err != nil {
		return nil, err
	}

	microVMs, err := pool.ListMicroVMs(ctx, poolName)
	if err != nil {
		return nil, err
	}

	return microVMs, nil
}

// GetMicroVM returns a MicroVM object by the given VM ID.
func (s *Server) GetMicroVM(ctx context.Context, vmid string) (*MicroVM, error) {
	s.l.Lock()
	pools := make([]*Pool, 0, len(s.pools))
	for _, pool := range s.pools {
		pools = append(pools, pool)
	}
	s.l.Unlock()

	for _, pool := range pools {
		pool.machinesMu.Lock()

		for _, metadata := range pool.machines {
			if metadata.machine.Cfg.VMID != vmid {
				continue
			}

			ip := metadata.machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String()
			vm := &MicroVM{VMID: vmid, Pool: pool.config.Name, IPAddr: ip, CreatedAt: metadata.createdAt}
			pool.machinesMu.Unlock()
			return vm, nil
		}

		pool.machinesMu.Unlock()
	}

	return nil, fmt.Errorf("machine with VMID %q not found", vmid)
}
