package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/errdefs"
	"github.com/containerd/log"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/hostinger/fireactions/helper/deepcopy"
	"github.com/hostinger/fireactions/helper/github"
	"github.com/hostinger/fireactions/helper/stringid"
	"github.com/opencontainers/image-spec/identity"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"

	githubv63 "github.com/google/go-github/v63/github"
)

const (
	defaultSnapshotter = "devmapper"
)

// machineMetadata holds metadata about a Firecracker machine and its associated resources.
type machineMetadata struct {
	machine     *firecracker.Machine
	runnerID    int64
	createdAt   time.Time
	leaseCancel func(context.Context) error // containerd lease cancel function
	logFile     *os.File
}

// Pool represents a pool of Firecracker VMs that are used to run GitHub Actions jobs.
type Pool struct {
	config         *PoolConfig
	containerd     *containerd.Client
	containerdMu   *sync.Mutex
	github         *github.Client
	imageManager   *imageManager
	machinesMu     *sync.Mutex
	machines       map[string]*machineMetadata // map[runnerName]machineMetadata
	installationID atomic.Int64                // GitHub App installation ID
	logger         *zerolog.Logger
	l              *sync.Mutex
	isActive       bool
	t              *time.Ticker
	stopCh         chan struct{}
	doneCh         chan struct{}  // Signals when Run() has exited
	cleanupWg      sync.WaitGroup // Tracks cleanup goroutines
	ctx            context.Context
	cancel         context.CancelFunc
}

// PoolConfig represents the configuration of a Pool.
type PoolConfig struct {
	Name        string             `yaml:"name" validate:"required"`
	Replicas    int                `yaml:"replicas" validate:"min=0"`
	Runner      *RunnerConfig      `yaml:"runner" validate:"required"`
	Firecracker *FirecrackerConfig `yaml:"firecracker" validate:"required"`
}

// NewPool creates a new Pool.
func NewPool(logger *zerolog.Logger, config *PoolConfig, github *github.Client, imageManager *imageManager, containerdClient *containerd.Client) (*Pool, error) {
	l := logger.With().Str("pool", config.Name).Logger()

	ctx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		config:       config,
		machinesMu:   &sync.Mutex{},
		machines:     make(map[string]*machineMetadata),
		isActive:     true,
		containerd:   containerdClient,
		containerdMu: &sync.Mutex{},
		github:       github,
		imageManager: imageManager,
		logger:       &l,
		l:            &sync.Mutex{},
		t:            time.NewTicker(1 * time.Second),
		stopCh:       make(chan struct{}, 1),
		doneCh:       make(chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
	}

	if _, err := os.Stat(p.GetDir()); os.IsNotExist(err) {
		if err := os.MkdirAll(p.GetDir(), 0755); err != nil {
			return nil, fmt.Errorf("creating pool directory: %w", err)
		}

		p.logger.Debug().Msgf("Pool directory created at %s", p.GetDir())
	}

	metricPoolRunnersCurrent.
		WithLabelValues(p.config.Name, p.config.Runner.Organization).Set(float64(p.GetCurrentSize()))
	metricPoolRunnersDesired.
		WithLabelValues(p.config.Name, p.config.Runner.Organization).Set(float64(p.config.Replicas))
	metricPoolStatus.
		WithLabelValues(p.config.Name).Set(1)

	metricPoolsTotal.Inc()

	return p, nil
}

// Run starts the pool. Starting the pool will start the scaling process.
func (p *Pool) Run() {
	defer p.t.Stop()
	defer close(p.doneCh) // Signal that Run() has exited
	for {
		select {
		case <-p.stopCh:
			return
		case <-p.ctx.Done():
			return
		case <-p.t.C:
		}

		curSize := p.GetCurrentSize()
		desiredReplicas := p.GetReplicas()
		metricPoolRunnersCurrent.
			WithLabelValues(p.config.Name, p.config.Runner.Organization).Set(float64(curSize))
		metricPoolRunnersDesired.
			WithLabelValues(p.config.Name, p.config.Runner.Organization).Set(float64(desiredReplicas))

		if !p.isActive {
			p.logger.Debug().Msgf("Pool %s is paused, skipping scaling", p.config.Name)
			continue
		}

		// Scale to desired replicas
		if err := p.Scale(p.ctx, desiredReplicas-p.GetCurrentSize()); err != nil {
			// Don't log errors if context was cancelled (pool is stopping)
			if p.ctx.Err() == nil {
				p.logger.Error().Err(err).Msg("Failed to scale pool")
			}
		}
	}
}

// Stop stops the pool. Stopping the pool will stop all the VMs in the pool.
func (p *Pool) Stop() {
	p.logger.Debug().Msgf("Stopping pool %s", p.config.Name)
	p.cancel()

	// Signal the Start() loop to exit (non-blocking)
	select {
	case p.stopCh <- struct{}{}:
	default:
		// Channel already has a value or Start() already exited
	}

	// Wait for Run() loop to exit cleanly with a timeout
	select {
	case <-p.doneCh:
	case <-time.After(5 * time.Second):
		p.logger.Warn().Msg("Timeout waiting for Run() to exit")
	}

	// Now acquire the lock to stop machines
	p.l.Lock()
	defer p.l.Unlock()

	p.logger.Debug().Msgf("Stopping %d machines in pool %s", len(p.machines), p.config.Name)

	p.machinesMu.Lock()
	metadataList := make([]*machineMetadata, 0, len(p.machines))
	for _, metadata := range p.machines {
		metadataList = append(metadataList, metadata)
	}
	p.machinesMu.Unlock()

	// Stop all machines - cleanup goroutines will handle the rest
	for _, metadata := range metadataList {
		runnerName := metadata.machine.Cfg.VMID

		err := metadata.machine.StopVMM()
		if err != nil {
			p.logger.Error().Err(err).Msgf("Failed to stop Firecracker VM %s", runnerName)
		}

		p.logger.Debug().Msgf("Stopped Firecracker VM %s", runnerName)
	}

	// Wait for all cleanup goroutines to finish with a timeout
	cleanupDone := make(chan struct{})
	go func() {
		p.cleanupWg.Wait()
		close(cleanupDone)
	}()

	select {
	case <-cleanupDone:
	case <-time.After(10 * time.Second):
		p.logger.Warn().Msg("Timeout waiting for cleanup goroutines to finish")
	}

	p.logger.Debug().Msgf("Pool %s stopped", p.config.Name)
}

// GetDir returns the directory where the pool sockets and logs are stored.
func (p *Pool) GetDir() string {
	return fmt.Sprintf("/var/lib/fireactions/pools/%s", p.config.Name)
}

// Scale scales the pool to the desired size.
func (p *Pool) Scale(ctx context.Context, replicas int) error {
	// Check if pool is shutting down before attempting to scale
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
	}

	p.l.Lock()
	defer p.l.Unlock()

	curSize := p.GetCurrentSize()
	desSize := curSize + replicas

	// If no change needed, return
	if desSize == curSize {
		return nil
	}

	// Scale up
	if replicas > 0 {
		for i := curSize; i < desSize; i++ {
			// Check context before each scale operation
			select {
			case <-p.ctx.Done():
				return p.ctx.Err()
			default:
			}

			start := time.Now()
			if err := p.scaleUp(ctx); err != nil {
				metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "up", "failure").Inc()
				return err
			}
			duration := time.Since(start).Seconds()

			metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "up", "success").Inc()
			metricScaleDuration.WithLabelValues(p.config.Name, p.config.Runner.Organization, "up").Observe(duration)
			p.logger.Trace().Msgf("Pool scaled to %d", i+1)
		}
	}

	// Scale down
	if replicas < 0 {
		toRemove := -replicas
		if toRemove > curSize {
			toRemove = curSize
		}

		for i := 0; i < toRemove; i++ {
			start := time.Now()
			if err := p.scaleDown(ctx); err != nil {
				metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "down", "failure").Inc()
				return err
			}
			duration := time.Since(start).Seconds()

			metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "down", "success").Inc()
			metricScaleDuration.WithLabelValues(p.config.Name, p.config.Runner.Organization, "down").Observe(duration)
		}
	}

	p.logger.Info().Msgf("Pool scaled %d -> %d (target: %d)", curSize, desSize, p.config.Replicas)
	return nil
}

// Pause pauses the pool. Pausing the pool will prevent the pool from scaling.
func (p *Pool) Pause() {
	if !p.isActive {
		return
	}

	p.logger.Debug().Msgf("Pool %s state changed to paused", p.config.Name)
	p.isActive = false
}

// Resume resumes the pool. Resuming the pool will allow the pool to scale.
func (p *Pool) Resume() {
	if p.isActive {
		return
	}

	p.logger.Debug().Msgf("Pool %s state changed to active", p.config.Name)
	p.isActive = true
}

// SetReplicas updates the desired replica count for the pool in a thread-safe manner.
func (p *Pool) SetReplicas(replicas int) {
	p.l.Lock()
	defer p.l.Unlock()
	p.config.Replicas = replicas
}

// GetReplicas returns the desired replica count for the pool in a thread-safe manner.
func (p *Pool) GetReplicas() int {
	p.l.Lock()
	defer p.l.Unlock()
	return p.config.Replicas
}

// GetCurrentSize returns the current size of the pool.
func (p *Pool) GetCurrentSize() int {
	p.machinesMu.Lock()
	defer p.machinesMu.Unlock()
	return len(p.machines)
}

func (p *Pool) scaleUp(ctx context.Context) error {
	// The containerd client already has the default namespace set, no need to wrap context

	// Use the shared image manager to get the image (handles deduplication and policy)
	image, err := p.imageManager.ensureImage(
		ctx,
		p.config.Runner.Image,
		p.config.Runner.ImagePullPolicy,
	)
	if err != nil {
		return fmt.Errorf("ensuring image: %w", err)
	}

	runnerName := fmt.Sprintf("%s-%s", p.config.Runner.Name, stringid.New())

	leaseCtx, leaseCtxCancel, err := p.containerd.WithLease(ctx,
		leases.WithID(fmt.Sprintf("fireactions/pools/%s/%s", p.config.Name, runnerName)))
	if err != nil {
		return fmt.Errorf("containerd: creating lease: %w", err)
	}

	// Track if we successfully created the machine to determine cleanup responsibility
	var machineCreated bool
	defer func() {
		if !machineCreated {
			// Clean up lease if machine creation failed
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cleanupCancel()
			_ = leaseCtxCancel(cleanupCtx)
		}
	}()

	snapshotMounts, err := p.createSnapshot(leaseCtx, image, runnerName)
	if err != nil {
		return fmt.Errorf("containerd: creating snapshot: %w", err)
	}

	machineLogFile, err := os.Create(filepath.Join(p.GetDir(), fmt.Sprintf("%s.log", runnerName)))
	if err != nil {
		return fmt.Errorf("creating log file: %w", err)
	}
	defer func() {
		if !machineCreated {
			// Close the file if we failed to create the machine
			_ = machineLogFile.Close()
		}
	}()

	machineCmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(filepath.Join(p.GetDir(), fmt.Sprintf("%s.sock", runnerName))).
		WithStderr(machineLogFile).
		WithStdout(machineLogFile).
		WithBin(p.config.Firecracker.BinaryPath).
		Build(context.Background())

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(io.Discard)

	machine, err := firecracker.NewMachine(ctx, firecracker.Config{
		VMID:            runnerName,
		SocketPath:      filepath.Join(p.GetDir(), fmt.Sprintf("%s.sock", runnerName)),
		KernelImagePath: p.config.Firecracker.KernelImagePath,
		KernelArgs:      p.config.Firecracker.KernelArgs,
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  &p.config.Firecracker.MachineConfig.VcpuCount,
			MemSizeMib: &p.config.Firecracker.MachineConfig.MemSizeMib,
		},
		Drives: []models.Drive{{
			DriveID:      firecracker.String("rootfs"),
			PathOnHost:   &snapshotMounts[0].Source,
			IsRootDevice: firecracker.Bool(true),
			IsReadOnly:   firecracker.Bool(false),
		}},
		NetworkInterfaces: []firecracker.NetworkInterface{{
			AllowMMDS:        true,
			CNIConfiguration: &firecracker.CNIConfiguration{NetworkName: "fireactions", IfName: "eth0", ConfDir: "/etc/cni/net.d", BinPath: []string{"/opt/cni/bin"}},
		}},
		MmdsAddress:    net.IPv4(169, 254, 169, 254),
		MmdsVersion:    firecracker.MMDSv2,
		ForwardSignals: []os.Signal{},
		LogPath:        filepath.Join(p.GetDir(), fmt.Sprintf("%s.firecracker.log", runnerName)),
		LogLevel:       "Debug",
	}, firecracker.WithProcessRunner(machineCmd), firecracker.WithLogger(logrus.NewEntry(logger)))
	if err != nil {
		return fmt.Errorf("firecracker: creating machine: %w", err)
	}

	installation, _, err := p.github.Apps.FindOrganizationInstallation(ctx, p.config.Runner.Organization)
	if err != nil {
		return fmt.Errorf("github: %w", err)
	}

	// Store installation ID for cleanup
	if p.installationID.Load() == 0 {
		p.installationID.Store(installation.GetID())
	}

	client := p.github.Installation(installation.GetID())
	jitConfig, _, err := client.Actions.GenerateOrgJITConfig(ctx, p.config.Runner.Organization, &githubv63.GenerateJITConfigRequest{
		Name:          runnerName,
		RunnerGroupID: p.config.Runner.GroupID,
		Labels:        p.config.Runner.Labels,
	})
	if err != nil {
		return fmt.Errorf("github: %w", err)
	}

	metadata := map[string]interface{}{"latest": map[string]interface{}{"meta-data": deepcopy.Map(p.config.Firecracker.Metadata)}}
	metadata["latest"].(map[string]interface{})["meta-data"].(map[string]interface{})["fireactions"] = map[string]interface{}{
		"runner_id":         runnerName,
		"runner_jit_config": jitConfig.GetEncodedJITConfig(),
	}

	machine.Handlers.FcInit = machine.Handlers.FcInit.Append(firecracker.NewSetMetadataHandler(metadata))

	if err := machine.Start(context.Background()); err != nil {
		return fmt.Errorf("firecracker: starting machine: %w", err)
	}

	// Mark machine as successfully created
	machineCreated = true

	p.logger.Info().Msgf("Successfully created Firecracker VM %s", runnerName)

	// Create machine metadata
	md := &machineMetadata{
		machine:     machine,
		runnerID:    jitConfig.Runner.GetID(),
		createdAt:   time.Now(),
		leaseCancel: leaseCtxCancel,
		logFile:     machineLogFile,
	}

	p.machinesMu.Lock()
	p.machines[runnerName] = md
	p.machinesMu.Unlock()

	// Start cleanup goroutine
	p.cleanupWg.Add(1)
	go func() {
		defer p.cleanupWg.Done()

		// Wait for machine to exit or pool context to be cancelled
		waitDone := make(chan struct{})
		go func() {
			_ = md.machine.Wait(context.Background())
			close(waitDone)
		}()

		select {
		case <-waitDone:
			// Machine exited normally
		case <-p.ctx.Done():
			// Pool is stopping, proceed with cleanup anyway
		}

		p.logger.Info().Msgf("Successfully cleaned up exited Firecracker VM %s", runnerName)

		p.machinesMu.Lock()
		metadata, exists := p.machines[runnerName]
		if exists {
			delete(p.machines, runnerName)
		}
		p.machinesMu.Unlock()

		// Clean up resources if metadata exists
		if exists && metadata != nil {
			// Delete GitHub runner
			p.deleteGitHubRunner(runnerName, metadata.runnerID)

			// Clean up containerd lease
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := metadata.leaseCancel(ctx)
			if err != nil && !errdefs.IsNotFound(err) {
				p.logger.Error().Err(err).Msgf("Failed to remove Containerd lease for Firecracker VM %s", runnerName)
			}

			// Close log file
			if metadata.logFile != nil {
				_ = metadata.logFile.Close()
			}
		}
	}()

	return nil
}

func (p *Pool) scaleDown(ctx context.Context) error {
	p.machinesMu.Lock()

	// Find a machine to remove (pick the first one)
	var targetMetadata *machineMetadata
	var targetName string
	for name, metadata := range p.machines {
		targetMetadata = metadata
		targetName = name
		break
	}

	if targetMetadata == nil {
		p.machinesMu.Unlock()
		return fmt.Errorf("no machines available to scale down")
	}

	p.machinesMu.Unlock()

	err := targetMetadata.machine.StopVMM()
	if err != nil {
		p.logger.Warn().Err(err).Msgf("Failed to gracefully stop Firecracker VM %s during scale down", targetName)
	}

	p.logger.Info().Msgf("Successfully scaled down Firecracker VM %s", targetName)
	return nil
}

func (p *Pool) createSnapshot(ctx context.Context, image containerd.Image, snapshotID string) ([]mount.Mount, error) {
	snapshotService := p.containerd.SnapshotService(defaultSnapshotter)
	snapshotExists := true
	_, err := snapshotService.Stat(ctx, snapshotID)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, err
		}

		snapshotExists = false
	}

	if !snapshotExists {
		if err := p.unpackImage(ctx, image); err != nil {
			return nil, fmt.Errorf("unpack: %w", err)
		}

		imageContent, err := image.RootFS(ctx)
		if err != nil {
			return nil, fmt.Errorf("image: rootfs: %w", err)
		}

		_, err = snapshotService.Prepare(ctx, snapshotID, identity.ChainID(imageContent).String())
		if err != nil {
			return nil, fmt.Errorf("prepare: %w", err)
		}
	}

	mounts, err := snapshotService.Mounts(ctx, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("mounts: %w", err)
	}

	return mounts, nil
}

func (p *Pool) unpackImage(ctx context.Context, image containerd.Image) error {
	isUnpacked, err := image.IsUnpacked(ctx, defaultSnapshotter)
	if err != nil {
		return err
	}

	if isUnpacked {
		return nil
	}

	return image.Unpack(ctx, defaultSnapshotter)
}

// deleteGitHubRunner removes a runner from GitHub Actions
func (p *Pool) deleteGitHubRunner(runnerName string, runnerID int64) {
	if runnerID == 0 {
		p.logger.Debug().Msgf("No GitHub runner ID found for %s, skipping deletion", runnerName)
		return
	}

	if p.installationID.Load() == 0 {
		p.logger.Warn().Msgf("No installation ID available, cannot delete runner %s", runnerName)
		return
	}

	client := p.github.Installation(p.installationID.Load())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Actions.RemoveOrganizationRunner(ctx, p.config.Runner.Organization, runnerID)
	if err != nil {
		p.logger.Error().Err(err).Msgf("Failed to delete GitHub runner %s (ID: %d)", runnerName, runnerID)
		return
	}

	p.logger.Info().Msgf("Successfully deleted GitHub runner %s (ID: %d)", runnerName, runnerID)
}

func init() {
	_ = log.SetLevel("panic")
}
