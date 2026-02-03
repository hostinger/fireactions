package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/containerd/containerd"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/helper/github"
	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents the Fireactions server.
type Server struct {
	serverv1.UnimplementedServerServiceServer
	config        *Config
	pools         map[string]*Pool
	grpcServer    *grpc.Server
	metricsServer *http.Server
	github        *github.Client
	containerd    *containerd.Client
	imageManager  *imageManager
	l             *sync.Mutex
	logger        *zerolog.Logger
	nextCID       atomic.Uint32 // Global VSOCK CID counter (starts at 3)
	version       string        // Version info for GetVersion RPC
	commit        string
	date          string
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

	logger := zerolog.Nop()

	containerdClient, err := containerd.New(config.Containerd.Address,
		containerd.WithTimeout(5*time.Second), containerd.WithDefaultNamespace(config.Containerd.Namespace))
	if err != nil {
		return nil, fmt.Errorf("containerd: creating client: %w", err)
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
	// TODO: Add auth interceptor for BasicAuth if config.BasicAuthEnabled
	// TODO: Add logging interceptor
	)

	s := &Server{
		config:     config,
		grpcServer: grpcServer,
		pools:      make(map[string]*Pool),
		github:     github,
		containerd: containerdClient,
		l:          &sync.Mutex{},
		logger:     &logger,
		version:    fireactions.Version,
		commit:     fireactions.Commit,
		date:       fireactions.Date,
	}

	// Initialize CID counter (CID 2 is reserved for host, start at 3)
	s.nextCID.Store(2)

	for _, opt := range opts {
		opt(s)
	}

	s.imageManager = newImageManager(s.logger, containerdClient)

	// Register gRPC service
	serverv1.RegisterServerServiceServer(grpcServer, s)

	// Enable reflection for grpcurl debugging
	reflection.Register(grpcServer)

	// Setup metrics server (HTTP) for Prometheus
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

	return s, nil
}

// Run starts the server and blocks until the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info().Str("version", fireactions.Version).Str("date", fireactions.Date).Str("commit", fireactions.Commit).Msgf("Starting gRPC server on %s", s.config.BindAddress)
	listener, err := net.Listen("tcp", s.config.BindAddress)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	for _, poolConfig := range s.config.Pools {
		pool, err := NewPool(s.logger, poolConfig, s.github, s.imageManager, s.containerd, &s.nextCID)
		if err != nil {
			return fmt.Errorf("creating pool: %w", err)
		}

		s.pools[poolConfig.Name] = pool
		go pool.Run()
		s.logger.Info().Msgf("Pool %s started", poolConfig.Name)
	}

	errGroup := &errgroup.Group{}
	errGroup.Go(func() error { return s.grpcServer.Serve(listener) })
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

		// Gracefully stop gRPC server
		s.grpcServer.GracefulStop()
	}()

	metricUp.Set(1)

	err = errGroup.Wait()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	s.logger.Info().Msg("Server stopped")
	return nil
}

func (s *Server) findPool(id string) (*Pool, error) {
	s.l.Lock()
	defer s.l.Unlock()

	pool, ok := s.pools[id]
	if !ok {
		return nil, fmt.Errorf("pool not found: %s", id)
	}

	return pool, nil
}

func (s *Server) findMachine(id string) (*Machine, error) {
	s.l.Lock()
	defer s.l.Unlock()

	for _, pool := range s.pools {
		machine, err := pool.GetMachine(id)
		if err == nil {
			return machine, nil
		}

		continue
	}

	return nil, fmt.Errorf("machine not found: %s", id)
}
