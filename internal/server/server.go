package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/internal/server/ghclient"
	"github.com/hostinger/fireactions/internal/server/scheduler"
	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/server/store/bbolt"
	"github.com/rs/zerolog"
)

// Server struct.
type Server struct {
	TLSConfig    *tls.Config
	Store        store.Store
	scheduler    *scheduler.Scheduler
	server       *http.Server
	ghClient     *ghclient.Client
	shutdownOnce sync.Once
	shutdownMu   sync.Mutex
	shutdownCh   chan struct{}
	log          zerolog.Logger
	cfg          *Config
}

// ServerOpt is a function that configures a Server.
type ServerOpt func(*Server)

// New creates a new Server.
func New(log zerolog.Logger, cfg *Config, opts ...ServerOpt) (*Server, error) {
	s := &Server{
		TLSConfig:    &tls.Config{},
		cfg:          cfg,
		shutdownOnce: sync.Once{},
		shutdownCh:   make(chan struct{}),
		shutdownMu:   sync.Mutex{},
		log:          log.With().Str("component", "server").Logger(),
	}

	ghClient, err := ghclient.New(&ghclient.Config{
		AppID:         cfg.GitHub.AppID,
		AppPrivateKey: cfg.GitHub.AppPrivateKey,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating GitHub client: %w", err)
	}
	s.ghClient = ghClient

	store, err := bbolt.New(filepath.Join(cfg.DataDir, "fireactions.db"))
	if err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}

	s.scheduler = scheduler.New(&s.log, cfg.Scheduler, store)
	s.Store = store

	for _, opt := range opts {
		opt(s)
	}

	mux := gin.New()
	mux.Use(requestid.New(requestid.WithCustomHeaderStrKey("X-Request-ID")))
	mux.Use(gin.Recovery())

	v1 := mux.Group("/api/v1")
	{
		v1.Handle(http.MethodDelete, "/jobs/:id", s.handleDelJob)
		v1.Handle(http.MethodGet, "/jobs", s.handleGetJobs)
		v1.Handle(http.MethodGet, "/jobs/:id", s.handleGetJob)
		v1.Handle(http.MethodGet, "/nodes", s.handleGetNodes)
		v1.Handle(http.MethodPost, "/nodes/:id/connect", s.handleNodeConnect)
		v1.Handle(http.MethodGet, "/nodes/:id", s.handleGetNode)
		v1.Handle(http.MethodPost, "/nodes/:id/disconnect", s.handleNodeDisconnect)
		v1.Handle(http.MethodPost, "/nodes", s.handleNodeRegister)
		v1.Handle(http.MethodPost, "/nodes/:id/runners/:runner/complete", s.handleCompleteRunnerAssignment)
		v1.Handle(http.MethodPost, "/nodes/:id/runners/:runner/accept", s.handleAcceptRunnerAssignment)
		v1.Handle(http.MethodPost, "/nodes/:id/runners/:runner/reject", s.handleRejectRunnerAssignment)
		v1.Handle(http.MethodDelete, "/nodes/:id", s.handleNodeDeregister)
		v1.Handle(http.MethodGet, "/nodes/:id/runners", s.handleGetNodeAssignments)
		v1.Handle(http.MethodPost, "/github/:organisation/registration-token", s.handleGitHubRegistrationToken)
		v1.Handle(http.MethodGet, "/runners/:id", s.handleGetRunner)
		v1.Handle(http.MethodGet, "/runners", s.handleGetRunners)
	}

	mux.Handle(http.MethodPost, "/webhook", s.handleGitHubWebhook)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	s.server = srv

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

		s.Store.Close()
		s.scheduler.Shutdown()

		close(s.shutdownCh)
	})
}

// Start starts the Server. It blocks until Shutdown() is called.
func (s *Server) Start() error {
	s.log.Info().Msgf("starting server on %s", s.server.Addr)

	err := s.scheduler.Start()
	if err != nil {
		return fmt.Errorf("error starting scheduler: %w", err)
	}

	err = s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}
