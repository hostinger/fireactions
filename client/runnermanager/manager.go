package runnermanager

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/hostinger/fireactions"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const defaultCNINetworkName = "fireactions"

type ImageManager interface {
	PullImage(
		ctx context.Context, imageRef string, imageOwner string) error
	ImageExists(
		ctx context.Context, imageRef string) (bool, error)
	CreateImageSnapshot(
		ctx context.Context, imageRef string, snapshotKey string) (string, error)
	DeleteSnapshot(
		ctx context.Context, snapshotKey string) error
}

// Config is the configuration for the Manager.
type Config struct {
	PollInterval         time.Duration
	FireactionsServerURL string
	FirecrackerConfig    *FirecrackerConfig
	CNIConfig            *CNIConfig
	StartTimeout         time.Duration
}

type FirecrackerConfig struct {
	BinaryPath      string
	SocketPath      string
	KernelImagePath string
	KernelArgs      string
	LogFilePath     string
	LogLevel        string
}

type CNIConfig struct {
	ConfDir string
	BinDirs []string
}

func (c *Config) Validate() error {
	var err error

	if c.PollInterval <= 0 {
		err = fmt.Errorf("poll-interval must be greater than 0")
	}

	return err
}

// Manager manages Firecracker VMs. It's responsible for polling the server for GitHub runners and
// reconciling them.
type Manager struct {
	config       *Config
	targetNodeID *string
	machines     map[string]*firecracker.Machine
	imageManager ImageManager
	attempts     map[string]int
	client       fireactions.Client
	stopCh       chan struct{}
	logger       *zerolog.Logger
	enabled      bool
	l            *sync.Mutex
}

// New creates a new Manager.
func New(
	logger *zerolog.Logger, client fireactions.Client, imageManager ImageManager, targetNodeID *string, config *Config,
) (*Manager, error) {
	if config == nil {
		config = &Config{PollInterval: 5 * time.Second}
	}

	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return newManager(logger, client, imageManager, targetNodeID, config), nil
}

func newManager(
	logger *zerolog.Logger, client fireactions.Client, imageManager ImageManager, targetNodeID *string, config *Config,
) *Manager {
	m := &Manager{
		config:       config,
		targetNodeID: targetNodeID,
		machines:     make(map[string]*firecracker.Machine),
		imageManager: imageManager,
		client:       client,
		stopCh:       make(chan struct{}),
		attempts:     make(map[string]int),
		logger:       logger,
		enabled:      true,
		l:            &sync.Mutex{},
	}

	return m
}

// Run starts the Manager. Blocks until the Manager is stopped via Stop().
func (m *Manager) Run() {
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

// Pause pauses the Manager.
func (m *Manager) Pause() {
	m.l.Lock()
	defer m.l.Unlock()

	if !m.enabled {
		return
	}

	m.enabled = false
}

// Resume resumes the Manager after it has been paused.
func (m *Manager) Resume() {
	m.l.Lock()
	defer m.l.Unlock()

	if m.enabled {
		return
	}

	m.enabled = true
}

// Poll polls the server for new GitHub runners.
func (m *Manager) Poll() {
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
				m.logger.Error().Err(err).Str("id", runner.ID).Msg("error reconciling runner")
			}
		}(runner)
	}

	wg.Wait()
}

func (m *Manager) Stop(ctx context.Context) error {
	close(m.stopCh)

	for _, machine := range m.machines {
		err := m.stopMachine(ctx, machine)
		if err != nil {
			return fmt.Errorf("error stopping Machine: %w", err)
		}

		m.logger.Info().Str("id", machine.Cfg.VMID).Msg("stopped runner")
	}

	return nil
}

func (m *Manager) reconcileRunner(ctx context.Context, runner *fireactions.Runner) error {
	switch runner.Status.Phase {
	case fireactions.RunnerPhasePending, fireactions.RunnerPhaseRunning:
		return m.ensureRunnerStarted(ctx, runner)
	case fireactions.RunnerPhaseCompleted, fireactions.RunnerPhaseFailed:
		return m.ensureRunnerStopped(ctx, runner)
	default:
		return fmt.Errorf("unknown runner lifecycle phase: %s", runner.Status.Phase)
	}
}

func (m *Manager) ensureRunnerStarted(ctx context.Context, runner *fireactions.Runner) error {
	machine, ok := m.machines[runner.ID]
	if ok {
		return nil
	}

	machine, err := m.createMachine(ctx, runner)
	if err != nil {
		return fmt.Errorf("error creating Machine: %w", err)
	}

	startCtx, cancel := context.WithTimeout(ctx, m.config.StartTimeout)
	defer cancel()

	start := time.Now()
	err = m.startMachine(startCtx, machine)
	if err != nil {
		if err := m.stopMachine(ctx, machine); err != nil {
			return fmt.Errorf("error stopping failed to start Firecracker VM: %w", err)
		}

		return fmt.Errorf("error starting Firecracker VM: %w", err)
	}

	m.machines[runner.ID] = machine
	m.logger.Info().Str("id", runner.ID).Msgf("started runner in %.3fs", time.Since(start).Seconds())

	return nil
}

func (m *Manager) ensureRunnerStopped(ctx context.Context, runner *fireactions.Runner) error {
	machine, ok := m.machines[runner.ID]
	if !ok {
		return nil
	}

	err := m.stopMachine(ctx, machine)
	if err != nil {
		return fmt.Errorf("error stopping Firecracker VM: %w", err)
	}

	err = m.imageManager.
		DeleteSnapshot(ctx, fmt.Sprintf("fireactions/%s", runner.ID))
	if err != nil {
		return fmt.Errorf("error deleting Firecracker VM snapshot: %w", err)
	}

	m.client.DeleteRunner(context.Background(), runner.ID)
	if err != nil {
		return fmt.Errorf("error deleting runner: %w", err)
	}

	delete(m.machines, runner.ID)
	m.logger.Info().Str("id", runner.ID).Msg("stopped runner")

	return nil
}

func (m *Manager) createMachine(ctx context.Context, runner *fireactions.Runner) (*firecracker.Machine, error) {
	config := firecracker.Config{
		VMID:            runner.ID,
		SocketPath:      fmt.Sprintf(m.config.FirecrackerConfig.SocketPath, runner.ID),
		KernelImagePath: m.config.FirecrackerConfig.KernelImagePath,
		KernelArgs:      m.config.FirecrackerConfig.KernelArgs,
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  &runner.Resources.VCPUs,
			MemSizeMib: firecracker.Int64(runner.Resources.MemoryBytes / 1024 / 1024),
		},
		NetworkInterfaces: []firecracker.NetworkInterface{{
			AllowMMDS:      true,
			InRateLimiter:  &models.RateLimiter{},
			OutRateLimiter: &models.RateLimiter{},
			CNIConfiguration: &firecracker.CNIConfiguration{
				NetworkName: defaultCNINetworkName,
				IfName:      "veth0",
				BinPath:     m.config.CNIConfig.BinDirs,
				ConfDir:     m.config.CNIConfig.ConfDir,
			},
		}},
		MmdsAddress: net.IPv4(169, 254, 169, 254),
		MmdsVersion: firecracker.MMDSv2,
		LogPath:     fmt.Sprintf(m.config.FirecrackerConfig.LogFilePath, runner.ID),
		LogLevel:    m.config.FirecrackerConfig.LogLevel,
	}

	cmd := firecracker.VMCommandBuilder{}.
		WithBin(m.config.FirecrackerConfig.BinaryPath).
		WithSocketPath(config.SocketPath).
		WithStdout(io.Discard).
		WithStderr(io.Discard).
		Build(context.Background())

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	machine, err := firecracker.NewMachine(
		context.Background(), config, firecracker.WithProcessRunner(cmd), firecracker.WithLogger(logrus.NewEntry(logger)))
	if err != nil {
		return nil, err
	}

	machine.Handlers.FcInit = machine.Handlers.FcInit.Prepend(
		m.newSetupRootDriveHandler(runner))
	machine.Handlers.FcInit = machine.Handlers.FcInit.Append(
		m.newSetMetadataHandler(runner))

	return machine, nil
}

func (m *Manager) startMachine(ctx context.Context, machine *firecracker.Machine) error {
	err := machine.Start(context.Background())
	if err != nil {
		return err
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}

		err := m.checkMachine(ctx, machine)
		if err != nil {
			continue
		}

		break
	}

	go func() {
		stopCtx := context.Background()
		errCh := make(chan error)
		go func() { errCh <- machine.Wait(stopCtx) }()
		select {
		case err := <-errCh:
			if err != nil && (strings.Contains(err.Error(), "signal: terminated") || strings.Contains(err.Error(), "signal: interrupt")) {
				delete(m.machines, machine.Cfg.VMID)
				return
			}

			m.logger.Warn().Err(err).Str("id", machine.Cfg.VMID).Msgf("unexpected Firecracker VM exit")
			delete(m.machines, machine.Cfg.VMID)
		case <-stopCtx.Done():
			return
		}
	}()

	return nil
}

func (m *Manager) stopMachine(ctx context.Context, machine *firecracker.Machine) error {
	_, err := machine.PID()
	if err != nil {
		return nil
	}

	defer func() {
		var err error

		err = os.Remove(machine.Cfg.SocketPath)
		if err != nil && !os.IsNotExist(err) {
			m.logger.Warn().Str("id", machine.Cfg.VMID).Err(err).Msg("error removing machine socket file")
		}

		err = os.Remove(fmt.Sprintf("%s/fireactions-%s", os.TempDir(), machine.Cfg.VMID))
		if err != nil && !os.IsNotExist(err) {
			m.logger.Warn().Str("id", machine.Cfg.VMID).Err(err).Msg("error removing machine temporary root drive mount directory")
		}
	}()

	err = machine.StopVMM()
	if err != nil {
		return err
	}

	err = machine.Wait(ctx)
	if err != nil && err == context.DeadlineExceeded {
		return err
	}

	return nil
}

func (m *Manager) checkMachine(ctx context.Context, machine *firecracker.Machine) error {
	ip := machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr

	certChecker := &ssh.CertChecker{
		IsHostAuthority: func(auth ssh.PublicKey, address string) bool { return true },
		HostKeyFallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		IsRevoked:       func(cert *ssh.Certificate) bool { return false },
	}

	sshConfig := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: certChecker.CheckHostKey,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", ip.IP.String()), sshConfig)
	if err != nil && !strings.Contains(err.Error(), "unable to authenticate") {
		return err
	}

	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	return nil
}
