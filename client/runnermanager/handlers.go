package runnermanager

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/leases"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/containerd"
	"github.com/hostinger/fireactions/client/rootfs"
)

func (m *Manager) newSetMetadataHandler(runner *fireactions.Runner) firecracker.Handler {
	metadata := map[string]interface{}{
		"fireactions-server-url": m.config.FireactionsServerURL,
		"fireactions-runner-id":  runner.ID,
	}

	return firecracker.NewSetMetadataHandler(metadata)
}

func (m *Manager) newSetupRootDriveHandler(runner *fireactions.Runner) firecracker.Handler {
	fn := func(ctx context.Context, machine *firecracker.Machine) error {
		switch runner.ImagePullPolicy {
		case fireactions.RunnerImagePullPolicyAlways:
			err := containerd.PullImage(ctx, m.containerd, runner.Image)
			if err != nil {
				return fmt.Errorf("containerd: pulling image: %w", err)
			}
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

		leaseID := fmt.Sprintf("fireactions/runner/%s", runner.ID)
		leaseCtx, err := containerd.NewContextWithLease(ctx, m.containerd, leaseID, leases.WithExpiration(6*time.Hour))
		if err != nil {
			return fmt.Errorf("containerd: creating context with lease: %w", err)
		}

		snapshotKey := fmt.Sprintf("fireactions/runner/%s", runner.ID)
		mounts, err := containerd.CreateSnapshot(leaseCtx, m.containerd, runner.Image, "devmapper", snapshotKey)
		if err != nil {
			return fmt.Errorf("containerd: creating snapshot: %w", err)
		}

		rootDrivePath := mounts[0].Source

		machine.Cfg.Drives = []models.Drive{{
			DriveID:      firecracker.String("rootfs"),
			PathOnHost:   firecracker.String(rootDrivePath),
			IsRootDevice: firecracker.Bool(true),
			IsReadOnly:   firecracker.Bool(false),
		}}

		rootfs, err := rootfs.New(rootDrivePath)
		if err != nil {
			return fmt.Errorf("rootfs: %w", err)
		}
		defer rootfs.Close()

		err = rootfs.SetupHostname(runner.Name)
		if err != nil {
			return fmt.Errorf("rootfs: setting up hostname: %w", err)
		}

		err = rootfs.SetupDNS()
		if err != nil {
			return fmt.Errorf("rootfs: setting up DNS: %w", err)
		}

		return nil
	}

	return firecracker.Handler{Name: "fcinit.SetupRootDrive", Fn: fn}
}
