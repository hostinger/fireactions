package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/errdefs"
	"github.com/containerd/log"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
	"github.com/distribution/reference"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/hostinger/fireactions/helper/deepcopy"
	"github.com/hostinger/fireactions/helper/github"
	"github.com/hostinger/fireactions/helper/stringid"
	"github.com/opencontainers/image-spec/identity"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"

	githubv62 "github.com/google/go-github/v62/github"
)

const (
	defaultSnapshotter = "devmapper"
)

// Pool represents a pool of Firecracker VMs that are used to run GitHub Actions jobs.
type Pool struct {
	config       *PoolConfig
	containerd   *containerd.Client
	containerdMu *sync.Mutex
	github       *github.Client
	machinesMu   *sync.Mutex
	machines     map[string]*firecracker.Machine
	logger       *zerolog.Logger
	l            *sync.Mutex
	isActive     bool
	t            *time.Ticker
	stopCh       chan struct{}
}

// PoolConfig represents the configuration of a Pool.
type PoolConfig struct {
	Name        string             `yaml:"name" validate:"required"`
	MaxRunners  int                `yaml:"max_runners" validate:"min=1"`
	MinRunners  int                `yaml:"min_runners" validate:"min=1"`
	Runner      *RunnerConfig      `yaml:"runner" validate:"required"`
	Firecracker *FirecrackerConfig `yaml:"firecracker" validate:"required"`
}

// NewPool creates a new Pool.
func NewPool(logger *zerolog.Logger, config *PoolConfig, github *github.Client) (*Pool, error) {
	l := logger.With().Str("pool", config.Name).Logger()
	containerd, err := containerd.New("/run/containerd/containerd.sock",
		containerd.WithDefaultNamespace(config.Name),
		containerd.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("containerd: creating client: %w", err)
	}

	p := &Pool{
		config:       config,
		machinesMu:   &sync.Mutex{},
		machines:     make(map[string]*firecracker.Machine),
		isActive:     true,
		containerd:   containerd,
		containerdMu: &sync.Mutex{},
		github:       github,
		logger:       &l,
		l:            &sync.Mutex{},
		t:            time.NewTicker(1 * time.Second),
		stopCh:       make(chan struct{}),
	}

	_, err = os.Stat(p.GetDir())
	if os.IsNotExist(err) {
		if err := os.MkdirAll(p.GetDir(), 0755); err != nil {
			return nil, fmt.Errorf("creating pool directory: %w", err)
		}

		p.logger.Debug().Msgf("Pool directory created at %s", p.GetDir())
	}

	metricPoolCurrentRunnersCount.
		WithLabelValues(p.config.Name).Set(float64(p.GetCurrentSize()))
	metricPoolMaxRunnersCount.
		WithLabelValues(p.config.Name).Set(float64(p.config.MaxRunners))
	metricPoolMinRunnersCount.
		WithLabelValues(p.config.Name).Set(float64(p.config.MinRunners))
	metricPoolStatus.
		WithLabelValues(p.config.Name).Set(1)

	metricPoolTotal.Inc()

	return p, nil
}

// Start starts the pool. Starting the pool will start the scaling process.
func (p *Pool) Start() {
	defer p.t.Stop()
	for {
		select {
		case <-p.stopCh:
			return
		case <-p.t.C:
		}

		metricPoolCurrentRunnersCount.WithLabelValues(p.config.Name).Set(float64(p.GetCurrentSize()))

		if !p.isActive {
			p.logger.Debug().Msgf("Pool %s is paused, skipping scaling", p.config.Name)
			continue
		}

		if err := p.Scale(context.Background(), p.config.MinRunners-p.GetCurrentSize()); err != nil {
			p.logger.Error().Err(err).Msg("Failed to scale pool")
		}
	}
}

// Stop stops the pool. Stopping the pool will stop all the VMs in the pool.
func (p *Pool) Stop() {
	p.stopCh <- struct{}{}
	p.logger.Debug().Msgf("Stopping pool %s", p.config.Name)
	p.t.Stop()
	p.l.Lock()
	defer p.l.Unlock()

	for _, machine := range p.machines {
		err := machine.StopVMM()
		if err != nil {
			p.logger.Error().Err(err).Msgf("Failed to stop Firecracker VM %s", machine.Cfg.VMID)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = machine.Wait(ctx)

		p.machinesMu.Lock()
		delete(p.machines, machine.Cfg.VMID)
		p.machinesMu.Unlock()

		p.logger.Debug().Msgf("Forcefully stopped Firecracker VM %s", machine.Cfg.VMID)
	}

	p.logger.Debug().Msgf("Pool %s stopped", p.config.Name)
}

// GetDir returns the directory where the pool sockets and logs are stored.
func (p *Pool) GetDir() string {
	return fmt.Sprintf("/var/lib/fireactions/pools/%s", p.config.Name)
}

// Scale scales the pool to the desired size.
func (p *Pool) Scale(ctx context.Context, replicas int) error {
	p.l.Lock()
	defer p.l.Unlock()

	if replicas < 0 {
		return nil
	}

	curSize := p.GetCurrentSize()
	desSize := curSize + replicas

	if desSize > p.config.MaxRunners || desSize < p.config.MinRunners || desSize == curSize {
		return nil
	}

	for i := curSize; i < desSize; i++ {
		if err := p.scaleUp(ctx); err != nil {
			metricPoolScaleFailures.WithLabelValues(p.config.Name).Inc()
			return err
		}

		metricPoolScaleSuccesses.WithLabelValues(p.config.Name).Inc()
		p.logger.Trace().Msgf("Pool scaled to %d", i+1)
	}

	p.logger.Debug().Msgf("Pool scaled %d -> %d (max: %d, min: %d)", curSize, desSize, p.config.MaxRunners, p.config.MinRunners)
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

// GetCurrentSize returns the current size of the pool.
func (p *Pool) GetCurrentSize() int {
	return len(p.machines)
}

func (p *Pool) scaleUp(ctx context.Context) error {
	imageExists := true
	image, err := p.containerd.GetImage(ctx, p.config.Runner.Image)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return fmt.Errorf("containerd: getting image: %w", err)
		}

		imageExists = false
	}

	if !imageExists {
		p.logger.Debug().Msg("Pulling image")

		start := time.Now()
		image, err = p.pullImage(ctx, p.config.Runner.Image)
		if err != nil {
			return fmt.Errorf("containerd: pulling image: %w", err)
		}

		p.logger.Debug().Msgf("Image pulled in %s", time.Since(start))
	}

	runnerName := fmt.Sprintf("%s-%s", p.config.Runner.Name, stringid.New())

	leaseCtx, leaseCtxCancel, err := p.containerd.WithLease(ctx,
		leases.WithID(fmt.Sprintf("fireactions/pools/%s/%s", p.config.Name, runnerName)))
	if err != nil {
		return fmt.Errorf("containerd: creating lease: %w", err)
	}

	snapshotMounts, err := p.createSnapshot(leaseCtx, image, runnerName)
	if err != nil {
		return fmt.Errorf("containerd: creating snapshot: %w", err)
	}

	machineLogFile, err := os.Create(filepath.Join(p.GetDir(), fmt.Sprintf("%s.log", runnerName)))
	if err != nil {
		return fmt.Errorf("creating log file: %w", err)
	}

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
	}, firecracker.WithProcessRunner(machineCmd), firecracker.WithLogger(logrus.NewEntry(logger)))
	if err != nil {
		return fmt.Errorf("firecracker: creating machine: %w", err)
	}

	installation, _, err := p.github.Apps.FindOrganizationInstallation(ctx, p.config.Runner.Organization)
	if err != nil {
		return fmt.Errorf("github: %w", err)
	}

	client := p.github.Installation(installation.GetID())
	jitConfig, _, err := client.Actions.GenerateOrgJITConfig(ctx, p.config.Runner.Organization, &githubv62.GenerateJITConfigRequest{
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

	go func() {
		_ = machine.Wait(context.Background())
		p.logger.Debug().Msgf("Firecracker VM %s exited", runnerName)

		p.machinesMu.Lock()
		delete(p.machines, runnerName)
		p.machinesMu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := leaseCtxCancel(ctx)
		if err != nil && !errdefs.IsNotFound(err) {
			p.logger.Error().Err(err).Msgf(`Failed to remove Containerd lease for Firecracker VM %s.
Run 'ctr --namespace %s leases rm fireactions/pools/%s/%s' to remove the lease manually`, runnerName, p.config.Name, p.config.Name, runnerName)
		}

		_ = machineLogFile.Close()
	}()

	if err := machine.Start(context.Background()); err != nil {
		return fmt.Errorf("firecracker: starting machine: %w", err)
	}

	p.logger.Debug().Msgf("Firecracker VM %s started", runnerName)
	p.machinesMu.Lock()
	p.machines[runnerName] = machine
	p.machinesMu.Unlock()

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

func (p *Pool) pullImage(ctx context.Context, ref string) (containerd.Image, error) {
	p.containerdMu.Lock()
	defer p.containerdMu.Unlock()

	image, err := p.containerd.GetImage(ctx, ref)
	if err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	} else if err == nil {
		return image, nil
	}

	dockerRef, err := reference.ParseDockerRef(ref)
	if err != nil {
		return nil, fmt.Errorf("parsing image ref: %w", err)
	}

	refDomain := reference.Domain(dockerRef)
	resolver, err := dockerconfigresolver.New(ctx, refDomain)
	if err != nil {
		return nil, fmt.Errorf("creating docker config resolver: %w", err)
	}

	image, err = p.containerd.Pull(ctx, ref,
		containerd.WithPullUnpack, containerd.WithResolver(resolver), containerd.WithPullSnapshotter(defaultSnapshotter))
	if err != nil {
		return nil, err
	}

	return image, nil
}

func init() {
	_ = log.SetLevel("panic")
}
