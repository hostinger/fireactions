package runnermanager

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/agent"
	"github.com/hostinger/fireactions/client/containerd"
	"github.com/hostinger/fireactions/client/firecracker"
	"github.com/rs/zerolog"
)

const defaultCNINetworkName = "fireactions"

// Config is the configuration for the Manager.
type Config struct {
	PollInterval               time.Duration
	StartTimeout               time.Duration
	FirecrackerBinaryPath      string
	FirecrackerKernelImagePath string
	FirecrackerKernelArgs      string
	FirecrackerLogLevel        string
	FirecrackerLogFilePath     string
	CNIConfDir                 string
	CNIBinDirs                 []string
}

type Manager interface {
	Run()
	Resume()
	Pause()
	Stop(ctx context.Context) error
}

type managerImpl struct {
	config       *Config
	targetNodeID *string
	machinesMu   sync.RWMutex
	machines     map[string]*firecracker.Machine
	client       fireactions.Client
	containerd   containerd.Client
	stopCh       chan struct{}
	logger       *zerolog.Logger
	enabled      bool
	l            *sync.Mutex
}

// New creates a new Manager implementation.
func New(logger *zerolog.Logger, client fireactions.Client, containerd containerd.Client, targetNodeID *string, config *Config) (*managerImpl, error) {
	if config == nil {
		config = &Config{PollInterval: 5 * time.Second}
	}

	m := &managerImpl{
		config:       config,
		targetNodeID: targetNodeID,
		machines:     make(map[string]*firecracker.Machine),
		machinesMu:   sync.RWMutex{},
		containerd:   containerd,
		client:       client,
		stopCh:       make(chan struct{}),
		logger:       logger,
		enabled:      true,
		l:            &sync.Mutex{},
	}

	return m, nil
}

func (m *managerImpl) Run() {
	t := time.NewTicker(m.config.PollInterval)
	defer t.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-t.C:
		}

		if !m.enabled {
			continue
		}

		m.Poll()
	}
}

func (m *managerImpl) Pause() {
	m.l.Lock()
	defer m.l.Unlock()

	if !m.enabled {
		return
	}

	m.enabled = false
}

// Resume resumes the Manager after it has been paused.
func (m *managerImpl) Resume() {
	m.l.Lock()
	defer m.l.Unlock()

	if m.enabled {
		return
	}

	m.enabled = true
}

// Poll polls the server for new GitHub runners.
func (m *managerImpl) Poll() {
	m.l.Lock()
	defer m.l.Unlock()

	runners, _, err := m.client.GetNodeRunners(context.Background(), *m.targetNodeID)
	if err != nil {
		m.logger.Error().Err(err).Msg("error getting runners")
		return
	}

	var wg sync.WaitGroup
	for _, runner := range runners {
		wg.Add(1)
		go func(runner *fireactions.Runner) {
			defer wg.Done()
			err := m.reconcileRunner(context.Background(), runner)
			if err != nil {
				m.logger.Error().Err(err).Str("id", runner.ID).Msg("failed to reconcile runner")
			}

			m.logger.Debug().Str("id", runner.ID).Msg("reconciled runner")
		}(runner)
	}

	wg.Wait()
}

func (m *managerImpl) Stop(ctx context.Context) error {
	close(m.stopCh)

	for _, machine := range m.machines {
		err := machine.Stop(ctx)
		if err != nil {
			return fmt.Errorf("machine %s: %w", machine.ID(), err)
		}

		m.delMachine(machine.ID())
	}

	return nil
}

func (m *managerImpl) reconcileRunner(ctx context.Context, runner *fireactions.Runner) error {
	switch runner.Status.Phase {
	case fireactions.RunnerPhasePending, fireactions.RunnerPhaseIdle, fireactions.RunnerPhaseActive:
		err := m.ensureRunnerStarted(ctx, runner)
		if err != nil {
			return fmt.Errorf("starting GitHub runner: %w", err)
		}
	case fireactions.RunnerPhaseCompleted:
		err := m.ensureRunnerStopped(ctx, runner)
		if err != nil {
			return fmt.Errorf("stopping GitHub runner: %w", err)
		}
	default:
		return fmt.Errorf("unknown runner status: %s", runner.Status.Phase)
	}

	return nil
}

func (m *managerImpl) ensureRunnerStarted(ctx context.Context, runner *fireactions.Runner) error {
	startCtx, cancel := context.WithTimeout(ctx, m.config.StartTimeout)
	defer cancel()

	return m.startRunner(startCtx, runner)
}

func (m *managerImpl) ensureRunnerStopped(ctx context.Context, runner *fireactions.Runner) error {
	err := m.stopRunner(ctx, runner)
	if err != nil {
		return err
	}

	_, err = m.client.DeleteRunner(context.Background(), runner.ID)
	if err != nil {
		return fmt.Errorf("client: deleting runner: %w", err)
	}

	return nil
}

func (m *managerImpl) stopRunner(ctx context.Context, runner *fireactions.Runner) error {
	machine, ok := m.getMachine(runner.ID)
	if !ok {
		return nil
	}

	if !machine.IsRunning() {
		return nil
	}

	runnerRemoveToken, _, err := m.client.GetRunnerRemoveToken(ctx, runner.ID)
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	agentClient := agent.NewClient(fmt.Sprintf("http://%s:6969", machine.Config().NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String()))
	_, err = agentClient.Stop(ctx, &agent.StopRequest{Token: runnerRemoveToken.Token})
	if err != nil {
		return fmt.Errorf("agent: %w", err)
	}
	defer agentClient.Close()

	err = machine.Stop(ctx)
	if err != nil {
		return fmt.Errorf("machine: %w", err)
	}

	err = containerd.DeleteLease(ctx, m.containerd, fmt.Sprintf("fireactions/runner/%s", runner.ID))
	if err != nil {
		return fmt.Errorf("containerd: deleting lease: %w", err)
	}

	m.delMachine(runner.ID)
	m.logger.Debug().Str("id", runner.ID).Msg("stopped runner")
	return nil
}

func (m *managerImpl) startRunner(ctx context.Context, runner *fireactions.Runner) error {
	logger := m.logger.With().Str("id", runner.ID).Logger()

	machine, ok := m.getMachine(runner.ID)
	if ok {
		return nil
	}

	switch runner.ImagePullPolicy {
	case fireactions.RunnerImagePullPolicyIfNotPresent:
		ok, err := containerd.ImageExists(ctx, m.containerd, runner.Image)
		if err != nil {
			return fmt.Errorf("containerd: getting image: %w", err)
		}

		if ok {
			break
		}

		err = containerd.PullImage(ctx, m.containerd, runner.Image)
		if err != nil {
			return fmt.Errorf("containerd: pulling image: %w", err)
		}
	case fireactions.RunnerImagePullPolicyAlways:
		err := containerd.PullImage(ctx, m.containerd, runner.Image)
		if err != nil {
			return fmt.Errorf("containerd: pulling image: %w", err)
		}
	case fireactions.RunnerImagePullPolicyNever:
		ok, err := containerd.ImageExists(ctx, m.containerd, runner.Image)
		if err != nil {
			return fmt.Errorf("containerd: getting image: %w", err)
		}

		if !ok {
			return fmt.Errorf("containerd: image not found")
		}
	default:
	}

	leaseCtx, err := containerd.NewContextWithLease(ctx, m.containerd, fmt.Sprintf("fireactions/runner/%s", runner.ID))
	if err != nil {
		return fmt.Errorf("containerd: creating context with lease: %w", err)
	}

	snapshotKey := fmt.Sprintf("fireactions/runner/%s", runner.ID)
	mounts, err := containerd.CreateSnapshot(leaseCtx, m.containerd, runner.Image, "devmapper", snapshotKey)
	if err != nil {
		return fmt.Errorf("containerd: creating snapshot: %w", err)
	}

	machineConfig, err := firecracker.NewConfigBuilder().
		WithID(runner.ID).
		WithSocketPath(fmt.Sprintf("/var/run/fireactions-%s.sock", uuid.New().String())).
		WithMemory(runner.Resources.MemoryBytes).
		WithVCPUs(runner.Resources.VCPUs).
		WithKernelImagePath(m.config.FirecrackerKernelImagePath).
		WithKernelArgs(m.config.FirecrackerKernelArgs).
		WithDrive("rootfs", mounts[0].Source, true, false).
		WithCNINetworkInterface(defaultCNINetworkName, "veth0", m.config.CNIConfDir, m.config.CNIBinDirs, true).
		WithLogLevel(m.config.FirecrackerLogLevel).
		WithLogPath(fmt.Sprintf(m.config.FirecrackerLogFilePath, runner.ID)).
		WithMetadata("fireactions", map[string]interface{}{
			"runner-id": runner.ID,
		}).
		WithMMDSAddress(net.IPv4(169, 254, 169, 254)).
		WithMMDSVersion("V2").
		Build()
	if err != nil {
		return err
	}

	machine = firecracker.NewMachine(machineConfig, firecracker.WithStderr(io.Discard), firecracker.WithStdout(io.Discard), firecracker.WithReadinessProbe(func(ctx context.Context, address string) error {
		client := agent.NewClient(fmt.Sprintf("http://%s:6969", address))

		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		_, err := client.Ping(ctx)
		if err != nil {
			return fmt.Errorf("agent: %w", err)
		}

		return nil
	}))

	err = machine.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting machine: %w", err)
	}

	m.setMachine(runner.ID, machine)
	go func() {
		err := <-machine.ExitCh()
		if err != nil && !strings.Contains(err.Error(), "signal: terminated") {
			logger.Error().Err(err).Msg("runner exited unexpectedly")
		}

		m.delMachine(runner.ID)
	}()

	runnerRegistrationToken, _, err := m.client.GetRunnerRegistrationToken(ctx, runner.ID)
	if err != nil {
		return fmt.Errorf("getting runner registration token: %w", err)
	}

	agentClient := agent.NewClient(fmt.Sprintf("http://%s:6969", machine.Config().NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String()))
	_, err = agentClient.Start(ctx, &agent.StartRequest{
		Name:          runner.Name,
		URL:           fmt.Sprintf("https://github.com/%s", runner.Organisation),
		Token:         runnerRegistrationToken.Token,
		Labels:        runner.Labels,
		Ephemeral:     true,
		DisableUpdate: true,
		Replace:       true,
	})
	if err != nil {
		return fmt.Errorf("starting agent: %w", err)
	}

	_, err = m.client.SetRunnerStatus(ctx, runner.ID, fireactions.SetRunnerStatusRequest{
		Phase: fireactions.RunnerPhaseIdle,
	})
	if err != nil {
		return fmt.Errorf("client: setting runner status: %w", err)
	}

	m.logger.Debug().Str("id", runner.ID).Msg("started runner successfully")
	return nil
}

func (m *managerImpl) getMachine(id string) (*firecracker.Machine, bool) {
	m.machinesMu.RLock()
	defer m.machinesMu.RUnlock()

	machine, ok := m.machines[id]
	return machine, ok
}

func (m *managerImpl) setMachine(id string, machine *firecracker.Machine) {
	m.machinesMu.Lock()
	defer m.machinesMu.Unlock()

	m.machines[id] = machine
}

func (m *managerImpl) delMachine(id string) {
	m.machinesMu.Lock()
	defer m.machinesMu.Unlock()

	delete(m.machines, id)
}
