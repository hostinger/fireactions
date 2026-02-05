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

// Pool represents a pool of Firecracker VMs that are used to run GitHub Actions jobs.
type Pool struct {
	config         *PoolConfig
	containerd     *containerd.Client
	github         *github.Client
	imageManager   *imageManager
	pendingCreates atomic.Int32
	pendingDeletes atomic.Int32
	machinesMu     *sync.Mutex
	machines       map[string]*Machine
	installationID atomic.Int64
	logger         *zerolog.Logger
	replicas       atomic.Int32
	isActive       bool
	scaleTrigger   chan struct{}
	stopCh         chan struct{}
	doneCh         chan struct{}
	cleanupWg      sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	nextCID        *atomic.Uint32
	l              *sync.Mutex
}

// PoolConfig represents the configuration of a Pool.
type PoolConfig struct {
	Name           string             `yaml:"name" validate:"required"`
	ShutdownOnExit *bool              `yaml:"shutdown_on_exit"`
	Replicas       int                `yaml:"replicas" validate:"min=0"`
	Runner         *RunnerConfig      `yaml:"runner" validate:"required"`
	Firecracker    *FirecrackerConfig `yaml:"firecracker" validate:"required"`
}

// UnmarshalYAML implements custom unmarshaling to set defaults.
func (p *PoolConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type poolConfigAlias PoolConfig
	defaults := poolConfigAlias{
		ShutdownOnExit: func() *bool { b := true; return &b }(),
	}

	if err := unmarshal(&defaults); err != nil {
		return err
	}

	*p = PoolConfig(defaults)
	return nil
}

// NewPool creates a new Pool.
func NewPool(logger *zerolog.Logger, config *PoolConfig, github *github.Client, imageManager *imageManager, containerdClient *containerd.Client, nextCID *atomic.Uint32) (*Pool, error) {
	l := logger.With().Str("pool", config.Name).Logger()

	ctx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		config:       config,
		l:            &sync.Mutex{},
		machinesMu:   &sync.Mutex{},
		machines:     make(map[string]*Machine),
		isActive:     true,
		containerd:   containerdClient,
		github:       github,
		imageManager: imageManager,
		logger:       &l,
		scaleTrigger: make(chan struct{}, 1),
		stopCh:       make(chan struct{}, 1),
		doneCh:       make(chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
		nextCID:      nextCID,
	}

	p.replicas.Store(int32(config.Replicas))

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
	defer close(p.doneCh) // Signal that Run() has exited

	// Trigger initial scale
	p.TriggerScale()

	for {
		select {
		case <-p.scaleTrigger:
		case <-time.After(2 * time.Second):
		case <-p.stopCh:
			return
		case <-p.ctx.Done():
			return
		}

		// Check if we should stop before scaling (non-blocking check)
		select {
		case <-p.ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
		}

		curSize := p.GetCurrentSize()
		desiredReplicas := p.GetReplicas()
		pendingCreates := int(p.pendingCreates.Load())
		pendingDeletes := int(p.pendingDeletes.Load())
		netPending := pendingCreates - pendingDeletes
		metricPoolRunnersCurrent.
			WithLabelValues(p.config.Name, p.config.Runner.Organization).Set(float64(curSize))
		metricPoolRunnersDesired.
			WithLabelValues(p.config.Name, p.config.Runner.Organization).Set(float64(desiredReplicas))
		metricPoolRunnersPending.
			WithLabelValues(p.config.Name, p.config.Runner.Organization).Set(float64(netPending))

		if !p.isActive {
			p.logger.Debug().Msgf("Pool %s is paused, skipping scaling", p.config.Name)
			continue
		}

		// Scale to desired replicas
		if err := p.Scale(p.ctx, desiredReplicas); err != nil {
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

	p.logger.Debug().Msgf("Stopping %d machines in pool %s", len(p.machines), p.config.Name)

	p.machinesMu.Lock()
	machines := make([]*Machine, 0, len(p.machines))
	for _, machine := range p.machines {
		machines = append(machines, machine)
	}
	p.machinesMu.Unlock()

	// Stop all machines - cleanup goroutines will handle the rest
	for _, machine := range machines {
		runnerName := machine.Cfg.VMID

		err := machine.StopVMM()
		if err != nil {
			p.logger.Error().Err(err).Msgf("Failed to stop Firecracker VM %s", runnerName)
		}

		p.logger.Debug().Msgf("Stopped Firecracker VM %s", runnerName)
	}

	cleanupDone := make(chan struct{})
	go func() {
		p.cleanupWg.Wait()
		close(cleanupDone)
	}()

	select {
	case <-cleanupDone:
	case <-time.After(35 * time.Second):
		p.logger.Warn().Msg("Timeout waiting for cleanup goroutines to finish")
	}

	p.logger.Debug().Msgf("Pool %s stopped", p.config.Name)
}

// GetDir returns the directory where the pool sockets and logs are stored.
func (p *Pool) GetDir() string {
	return fmt.Sprintf("/var/lib/fireactions/pools/%s", p.config.Name)
}

// Scale scales the pool to the desired size.
func (p *Pool) Scale(ctx context.Context, desiredReplicas int) error {
	p.l.Lock()
	defer p.l.Unlock()

	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
	}

	curSize := p.GetCurrentSize()
	pendingCreates := int(p.pendingCreates.Load())
	pendingDeletes := int(p.pendingDeletes.Load())

	// Calculate effective size accounting for in-flight operations
	effectiveSize := curSize + pendingCreates - pendingDeletes
	delta := desiredReplicas - effectiveSize

	if delta == 0 {
		return nil
	}

	if delta > 0 {
		p.scaleUp(
			ctx, delta, desiredReplicas, curSize, pendingCreates, pendingDeletes)
	} else {
		p.scaleDown(
			ctx, -delta, desiredReplicas, curSize, pendingCreates, pendingDeletes)
	}

	return nil
}

func (p *Pool) scaleUp(ctx context.Context, count, desiredReplicas, curSize, pendingCreates, pendingRemovals int) {
	p.logger.Debug().Msgf("Scaling up by %d VMs (target: %d, current: %d, pending creates: %d, pending removals: %d)",
		count, desiredReplicas, curSize, pendingCreates, pendingRemovals)

	for i := 0; i < count; i++ {
		p.pendingCreates.Add(1)

		go func() {
			defer p.pendingCreates.Add(-1)

			select {
			case <-p.ctx.Done():
				return
			default:
			}

			start := time.Now()
			if err := p.createMachine(ctx); err != nil {
				metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "up", "failure").Inc()
				p.logger.Error().Err(err).Msg("Failed to create machine")
				return
			}

			duration := time.Since(start).Seconds()
			metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "up", "success").Inc()
			metricScaleDuration.WithLabelValues(p.config.Name, p.config.Runner.Organization, "up").Observe(duration)
		}()
	}
}

func (p *Pool) scaleDown(ctx context.Context, count, desiredReplicas, curSize, pendingCreates, pendingDeletes int) {
	if count > curSize {
		count = curSize
	}

	p.logger.Debug().Msgf("Scaling down by %d VMs (target: %d, current: %d, pending creates: %d, pending deletes: %d)",
		count, desiredReplicas, curSize, pendingCreates, pendingDeletes)

	for i := 0; i < count; i++ {
		p.pendingDeletes.Add(1)

		go func() {
			defer p.pendingDeletes.Add(-1)

			select {
			case <-p.ctx.Done():
				return
			default:
			}

			start := time.Now()
			if err := p.deleteMachine(ctx); err != nil {
				metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "down", "failure").Inc()
				p.logger.Error().Err(err).Msg("Failed to delete machine")
				return
			}

			duration := time.Since(start).Seconds()
			metricScaleOperations.WithLabelValues(p.config.Name, p.config.Runner.Organization, "down", "success").Inc()
			metricScaleDuration.WithLabelValues(p.config.Name, p.config.Runner.Organization, "down").Observe(duration)
		}()
	}
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
	p.replicas.Store(int32(replicas))
	p.TriggerScale()
}

// TriggerScale sends a non-blocking notification to trigger scaling.
func (p *Pool) TriggerScale() {
	select {
	case p.scaleTrigger <- struct{}{}:
	default:
	}
}

// GetReplicas returns the desired replica count for the pool in a thread-safe manner.
func (p *Pool) GetReplicas() int {
	return int(p.replicas.Load())
}

// GetCurrentSize returns the current size of the pool.
func (p *Pool) GetCurrentSize() int {
	p.machinesMu.Lock()
	defer p.machinesMu.Unlock()
	return len(p.machines)
}

func (p *Pool) ListMachines(ctx context.Context) ([]*Machine, error) {
	p.machinesMu.Lock()
	defer p.machinesMu.Unlock()

	machines := make([]*Machine, 0, len(p.machines))
	for _, machine := range p.machines {
		machines = append(machines, machine)
	}

	return machines, nil
}

func (p *Pool) GetMachine(name string) (*Machine, error) {
	p.machinesMu.Lock()
	defer p.machinesMu.Unlock()

	machine, ok := p.machines[name]
	if !ok {
		return nil, fmt.Errorf("machine not found: %s", name)
	}

	return machine, nil
}

func (p *Pool) createMachine(ctx context.Context) error {
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
	defer machineLogFile.Close()

	machineCmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(filepath.Join(p.GetDir(), fmt.Sprintf("%s.sock", runnerName))).
		WithStderr(machineLogFile).
		WithStdout(machineLogFile).
		WithBin(p.config.Firecracker.BinaryPath).
		Build(ctx)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(io.Discard)

	vsockPath := filepath.Join(p.GetDir(), fmt.Sprintf("%s.vsock", runnerName))
	vsockCID := p.nextCID.Add(1)

	fcMachine, err := firecracker.NewMachine(ctx, firecracker.Config{
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
		VsockDevices:   []firecracker.VsockDevice{{Path: vsockPath, CID: vsockCID}},
		MmdsAddress:    net.IPv4(169, 254, 169, 254),
		MmdsVersion:    firecracker.MMDSv2,
		ForwardSignals: []os.Signal{},
		LogPath:        filepath.Join(p.GetDir(), fmt.Sprintf("%s.firecracker.log", runnerName)),
		LogLevel:       "Debug",
	}, firecracker.WithProcessRunner(machineCmd), firecracker.WithLogger(logrus.NewEntry(logger)))
	if err != nil {
		return fmt.Errorf("firecracker: creating machine: %w", err)
	}

	installationID := p.installationID.Load()
	if installationID == 0 {
		installation, _, err := p.github.Apps.FindOrganizationInstallation(ctx, p.config.Runner.Organization)
		if err != nil {
			return fmt.Errorf("github: %w", err)
		}
		installationID = installation.GetID()

		p.installationID.Store(installationID)
	}

	client := p.github.Installation(installationID)
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
		"hostname":          runnerName,
		"shutdown_on_exit":  *p.config.ShutdownOnExit,
	}

	fcMachine.Handlers.FcInit = fcMachine.Handlers.FcInit.Append(firecracker.NewSetMetadataHandler(metadata))

	vmmCtx, vmmCancel := context.WithCancel(p.ctx)
	if err := fcMachine.Start(vmmCtx); err != nil {
		vmmCancel()
		return fmt.Errorf("firecracker: starting machine: %w", err)
	}

	// Mark machine as successfully created
	machineCreated = true

	p.logger.Info().Msgf("Successfully created Firecracker VM %s", runnerName)

	machine := &Machine{
		Machine:     fcMachine,
		Name:        jitConfig.GetRunner().GetName(),
		RunnerID:    jitConfig.GetRunner().GetID(),
		Pool:        p.config.Name,
		CreatedAt:   time.Now().UTC(),
		vsockCID:    vsockCID,
		vsockPath:   vsockPath,
		leaseCancel: leaseCtxCancel,
		vmmCtx:      vmmCtx,
		vmmCancel:   vmmCancel,
	}

	p.machinesMu.Lock()
	p.machines[runnerName] = machine
	p.machinesMu.Unlock()

	// Start cleanup goroutine
	p.cleanupWg.Add(1)
	go func() {
		defer p.cleanupWg.Done()

		waitDone := make(chan error, 1)
		go func() {
			waitDone <- machine.Wait(context.Background())
		}()

		select {
		case <-waitDone:
			// Machine exited normally
		case <-p.ctx.Done():
			// Pool is stopping, wait up to 30s for machine to fully exit
			select {
			case <-waitDone:
			case <-time.After(30 * time.Second):
				p.logger.Warn().Msgf("Timeout waiting for machine %s to exit during pool shutdown", runnerName)
			}
		}

		p.machinesMu.Lock()
		_, exists := p.machines[runnerName]
		if exists {
			delete(p.machines, runnerName)
		}
		p.machinesMu.Unlock()

		machine.vmmCancel()

		p.deleteGitHubRunner(runnerName, machine.RunnerID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := machine.leaseCancel(ctx)
		if err != nil && !errdefs.IsNotFound(err) {
			p.logger.Error().Err(err).Msgf("Failed to remove Containerd lease for Firecracker VM %s", runnerName)
		}

		p.logger.Info().Msgf("Successfully cleaned up exited Firecracker VM %s", runnerName)
	}()

	return nil
}

// removeMachine removes a single machine from the pool.
func (p *Pool) deleteMachine(_ context.Context) error {
	p.machinesMu.Lock()

	// Find a machine to remove (pick the first one)
	var targetMachine *Machine
	var targetName string
	for name, machine := range p.machines {
		targetMachine = machine
		targetName = name
		break
	}

	if targetMachine == nil {
		p.machinesMu.Unlock()
		return fmt.Errorf("no machines available to scale down")
	}

	// Remove from map immediately to prevent selecting the same machine multiple times
	delete(p.machines, targetName)
	p.machinesMu.Unlock()

	err := targetMachine.StopVMM()
	if err != nil {
		p.logger.Warn().Err(err).Msgf("Failed to stop VM %s", targetName)
		return err
	}

	p.logger.Info().Msgf("Successfully removed VM %s", targetName)
	return nil
}

// createSnapshot creates a snapshot of the specified image.
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

	p.logger.Debug().Msgf("Successfully deleted GitHub runner %s (ID: %d)", runnerName, runnerID)
}

func init() {
	_ = log.SetLevel("panic")
}
