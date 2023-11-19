package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/containerd"
	"github.com/hostinger/fireactions/client/microvm"
	"github.com/hostinger/fireactions/pkg/planner"
	"github.com/rs/zerolog"
)

// Config is the configuration for the Manager.
type Config struct {
	PollInterval time.Duration
	NodeID       *string
}

// Manager is the interface for the Manager. It is responsible for polling the Fireactions server
// for new GitHub runners and reconciling them.
type Manager interface {
	Run()
	Stop(ctx context.Context) error
}

type managerImpl struct {
	config     *Config
	driver     microvm.Driver
	client     fireactions.Client
	logger     *zerolog.Logger
	stopCh     chan struct{}
	containerd containerd.Client
	planner    planner.Planner
	wg         sync.WaitGroup
}

// New creates a new Manager.
func New(
	logger *zerolog.Logger, client fireactions.Client, containerd containerd.Client, driver microvm.Driver, config *Config) *managerImpl {
	manager := &managerImpl{
		config:     config,
		driver:     driver,
		client:     client,
		logger:     logger,
		stopCh:     make(chan struct{}),
		containerd: containerd,
		planner:    planner.NewPlanner(),
		wg:         sync.WaitGroup{},
	}

	return manager
}

func (m *managerImpl) Run() {
	go m.runPollingLoop()
}

func (m *managerImpl) Stop(ctx context.Context) error {
	m.stopCh <- struct{}{}
	return nil
}

func (m *managerImpl) runPollingLoop() {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-m.stopCh:
			t.Stop()
			return
		case <-t.C:
		}

		err := m.Poll(context.Background())
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to poll")
		}

		m.wg.Wait()
	}
}

func (m *managerImpl) Poll(ctx context.Context) error {
	runners, _, err := m.client.GetNodeRunners(ctx, *m.config.NodeID)
	if err != nil {
		return fmt.Errorf("client: getting node runners: %w", err)
	}

	for _, runner := range runners {
		m.wg.Add(1)
		go m.reconcileRunner(ctx, runner)
	}

	return nil
}

func (m *managerImpl) reconcileRunner(ctx context.Context, runner *fireactions.Runner) {
	defer m.wg.Done()

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var err error
	switch runner.Status.Phase {
	case fireactions.RunnerPhasePending, fireactions.RunnerPhaseIdle, fireactions.RunnerPhaseActive:
		err = m.planner.Execute(ctx, newCreateGitHubRunnerPlan(m.containerd, m.client, m.driver, runner))
	case fireactions.RunnerPhaseCompleted:
		err = m.planner.Execute(ctx, newDeleteGitHubRunnerPlan(m.containerd, m.client, m.driver, runner))
	default:
	}

	if err != nil {
		m.logger.Error().Err(err).Str("id", runner.ID).Msg("failed to reconcile runner")
		return
	}

	m.logger.Info().Str("id", runner.ID).Msg("reconciled runner")
}
