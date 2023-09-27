package runtime

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/log"
	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/firecracker"
	"github.com/hostinger/fireactions/client/microvm"
	"github.com/opencontainers/image-spec/identity"
	"github.com/rs/zerolog"
)

const (
	defaultSnapshotter = "devmapper"
)

// Config is the configuration for the Runtime.
type Config struct {
	ContainerdAddress string                    `mapstructure:"containerd_address"`
	Firecracker       *firecracker.DriverConfig `mapstructure:"firecracker"`
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	c := &Config{
		ContainerdAddress: "/run/containerd/containerd.sock",
		Firecracker:       firecracker.NewDriverConfig(),
	}

	return c
}

// Validate validates the Config.
func (c *Config) Validate() error {
	var errs error

	if c.ContainerdAddress == "" {
		errs = multierror.Append(errs, fmt.Errorf("containerd_address is required"))
	}

	if c.Firecracker == nil {
		errs = multierror.Append(errs, fmt.Errorf("driver_config is required"))
	} else {
		if err := c.Firecracker.Validate(); err != nil {
			errs = fmt.Errorf("driver_config: %w", err)
		}
	}

	return errs
}

// Runtime is the interface for the Runtime.
type Runtime interface {
	Sync(ctx context.Context, runner *fireactions.Runner) error
	Shutdown(ctx context.Context) error
}

type runtimeImpl struct {
	containerd            *containerd.Client
	containerdImagePuller *containerdImagePuller
	microvmManager        microvm.Manager
	client                fireactions.Client
	clientID              string
	logger                *zerolog.Logger
}

// NewRuntime creates a new Runtime.
func NewRuntime(logger *zerolog.Logger, client fireactions.Client, clientID string, config *Config) (*runtimeImpl, error) {
	containerd, err := containerd.New(config.ContainerdAddress, containerd.WithDefaultNamespace("fireactions"))
	if err != nil {
		return nil, fmt.Errorf("containerd: %w", err)
	}

	runtime := &runtimeImpl{
		containerd:            containerd,
		containerdImagePuller: newContainerdImagePuller(containerd),
		microvmManager:        microvm.NewInMemoryManager(firecracker.NewDriver(config.Firecracker)),
		client:                client,
		clientID:              clientID,
		logger:                logger,
	}

	return runtime, nil
}

func (r *runtimeImpl) Sync(ctx context.Context, runner *fireactions.Runner) error {
	switch runner.Status.State {
	case fireactions.RunnerStateCompleted:
		return r.stopRunner(ctx, runner)
	case fireactions.RunnerStatePending:
		return r.startRunner(ctx, runner)
	}

	return nil
}

func (r *runtimeImpl) List(ctx context.Context) ([]*fireactions.Runner, error) {
	runners, _, err := r.client.GetNodeRunners(ctx, r.clientID)
	if err != nil {
		return nil, err
	}

	return runners, nil
}

func (r *runtimeImpl) Shutdown(ctx context.Context) error {
	microvms, err := r.microvmManager.ListMicroVMs(ctx)
	if err != nil {
		return fmt.Errorf("microvm: %w", err)
	}

	for _, microvm := range microvms {
		err = r.microvmManager.StopMicroVM(ctx, microvm.Spec.ID)
		if err != nil {
			return fmt.Errorf("microvm: %w", err)
		}

		r.logger.Info().Msgf("stopped Micro VM: %s", microvm.Spec.ID)
	}

	return r.containerd.Close()
}

func (r *runtimeImpl) startRunner(ctx context.Context, runner *fireactions.Runner) error {
	imageExists := true
	image, err := r.getContainerImage(ctx, runner.Image)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return fmt.Errorf("containerd: getting container image: %w", err)
		}

		imageExists = false
	}

	if !imageExists && runner.ImagePullPolicy == fireactions.RunnerImagePullPolicyNever {
		return fmt.Errorf("containerd: container image not found: %s", runner.Image)
	}

	if !imageExists || runner.ImagePullPolicy == fireactions.RunnerImagePullPolicyAlways {
		image, err = r.containerdImagePuller.Pull(ctx, runner.Image)
		if err != nil {
			return fmt.Errorf("containerd: pulling container image: %w", err)
		}
	}

	mounts, err := r.createContainerImageSnapshot(ctx, image, fmt.Sprintf("fireactions/runner/%s", runner.ID))
	if err != nil {
		return fmt.Errorf("containerd: %w", err)
	}

	microvm, err := r.microvmManager.CreateMicroVM(ctx, microvm.Spec{
		ID:                runner.ID,
		Name:              runner.ID,
		VCPU:              runner.Resources.VCPUs,
		MemoryBytes:       runner.Resources.MemoryMB * 1024 * 1024,
		Metadata:          runner.Metadata,
		Drives:            []microvm.Drive{{PathOnHost: mounts[0].Source, ID: "rootfs", IsReadOnly: false, IsRoot: true}},
		NetworkInterfaces: []microvm.NetworkInterface{{AllowMMDS: true, NetworkName: "fireactions", IfName: "eth0"}},
	})
	if err != nil {
		return fmt.Errorf("microvm: %w", err)
	}

	err = r.microvmManager.StartMicroVM(ctx, microvm.Spec.ID)
	if err != nil {
		return fmt.Errorf("microvm: %w", err)
	}

	_, err = r.client.SetRunnerStatus(ctx, runner.ID, fireactions.SetRunnerStatusRequest{
		State:       fireactions.RunnerStateIdle,
		Description: "",
	})
	if err != nil {
		return fmt.Errorf("client: setting runner status: %w", err)
	}

	return nil
}

func (r *runtimeImpl) stopRunner(ctx context.Context, runner *fireactions.Runner) error {
	err := r.microvmManager.DeleteMicroVM(ctx, runner.ID)
	if err != nil && err != microvm.ErrNotFound {
		return fmt.Errorf("microvm: %w", err)
	}

	err = r.deleteContainerImageSnapshot(ctx, runner)
	if err != nil {
		return fmt.Errorf("containerd: %w", err)
	}

	_, err = r.client.DeleteRunner(ctx, runner.ID)
	if err != nil {
		return fmt.Errorf("client: deleting runner: %w", err)
	}

	return nil
}

func (r *runtimeImpl) getContainerImage(ctx context.Context, imageRef string) (containerd.Image, error) {
	image, err := r.containerd.GetImage(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (r *runtimeImpl) unpackContainerImage(ctx context.Context, image containerd.Image) error {
	isUnpacked, err := image.IsUnpacked(ctx, defaultSnapshotter)
	if err != nil {
		return err
	}

	if isUnpacked {
		return nil
	}

	return image.Unpack(ctx, defaultSnapshotter)
}

func (r *runtimeImpl) createContainerImageSnapshot(ctx context.Context, image containerd.Image, snapshotID string) ([]mount.Mount, error) {
	lease, err := r.createContainerImageSnapshotLease(ctx, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("creating lease: %w", err)
	}

	ctx = leases.WithLease(ctx, lease.ID)

	snapshotService := r.containerd.SnapshotService(defaultSnapshotter)
	snapshotExists := true
	_, err = snapshotService.Stat(ctx, snapshotID)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, err
		}

		snapshotExists = false
	}

	var mounts []mount.Mount
	if !snapshotExists {
		err = r.unpackContainerImage(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("unpack: %w", err)
		}

		imageContent, err := image.RootFS(ctx)
		if err != nil {
			return nil, fmt.Errorf("image: rootfs: %w", err)
		}

		mounts, err = snapshotService.Prepare(ctx, snapshotID, identity.ChainID(imageContent).String())
		if err != nil {
			return nil, fmt.Errorf("prepare: %w", err)
		}
	} else {
		mounts, err = snapshotService.Mounts(ctx, snapshotID)
		if err != nil {
			return nil, fmt.Errorf("mounts: %w", err)
		}
	}

	return mounts, nil
}

func (r *runtimeImpl) deleteContainerImageSnapshot(ctx context.Context, runner *fireactions.Runner) error {
	err := r.deleteContainerImageSnapshotLease(ctx, fmt.Sprintf("fireactions/runner/%s", runner.ID))
	if err != nil {
		return fmt.Errorf("deleting lease: %w", err)
	}

	snapshotService := r.containerd.SnapshotService(defaultSnapshotter)
	snapshotExists := true
	_, err = snapshotService.Stat(ctx, fmt.Sprintf("fireactions/runner/%s", runner.ID))
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return err
		}

		snapshotExists = false
	}

	if snapshotExists {
		err = snapshotService.Remove(ctx, fmt.Sprintf("fireactions/runner/%s", runner.ID))
		if err != nil {
			return fmt.Errorf("removing snapshot: %w", err)
		}
	}

	return nil
}

func (r *runtimeImpl) createContainerImageSnapshotLease(ctx context.Context, leaseID string) (*leases.Lease, error) {
	leasesService := r.containerd.LeasesService()
	existingLeases, err := leasesService.List(ctx, fmt.Sprintf("id==%s", leaseID))
	if err != nil {
		return nil, err
	}

	for _, lease := range existingLeases {
		if lease.ID != leaseID {
			continue
		}

		return &lease, nil
	}

	lease, err := leasesService.Create(ctx, leases.WithID(leaseID))
	if err != nil {
		return nil, err
	}

	return &lease, nil
}

func (r *runtimeImpl) deleteContainerImageSnapshotLease(ctx context.Context, leaseID string) error {
	leasesService := r.containerd.LeasesService()
	existingLeases, err := leasesService.List(ctx, fmt.Sprintf("id==%s", leaseID))
	if err != nil {
		return err
	}

	for _, lease := range existingLeases {
		if lease.ID != leaseID {
			continue
		}

		return leasesService.Delete(ctx, lease)
	}

	return nil
}

func init() {
	log.SetLevel("panic")
}
