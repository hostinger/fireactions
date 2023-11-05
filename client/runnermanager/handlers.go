package runnermanager

import (
	"context"
	"fmt"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/hostinger/fireactions"
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
			err := m.imageManager.PullImage(ctx, runner.Image, runner.ID)
			if err != nil {
				return fmt.Errorf("error pulling image: %w", err)
			}
		case fireactions.RunnerImagePullPolicyIfNotPresent:
			ok, err := m.imageManager.ImageExists(ctx, runner.Image)
			if err != nil {
				return fmt.Errorf("error checking if image exists: %w", err)
			}

			if ok {
				break
			}

			err = m.imageManager.PullImage(ctx, runner.Image, runner.ID)
			if err != nil {
				return fmt.Errorf("error pulling image: %w", err)
			}
		case fireactions.RunnerImagePullPolicyNever:
		default:
		}

		rootDrivePath, err := m.imageManager.CreateImageSnapshot(
			ctx, runner.Image, fmt.Sprintf("fireactions/%s", runner.ID))
		if err != nil {
			return fmt.Errorf("error creating image snapshot: %w", err)
		}

		machine.Cfg.Drives = []models.Drive{{
			DriveID:      firecracker.String("rootfs"),
			PathOnHost:   firecracker.String(rootDrivePath),
			IsRootDevice: firecracker.Bool(true),
			IsReadOnly:   firecracker.Bool(false),
		}}

		rootfs, err := rootfs.New(rootDrivePath)
		if err != nil {
			return fmt.Errorf("error creating rootfs: %w", err)
		}
		defer rootfs.Close()

		err = rootfs.SetupHostname(runner.Name)
		if err != nil {
			return fmt.Errorf("error setting up hostname: %w", err)
		}

		err = rootfs.SetupDNS()
		if err != nil {
			return fmt.Errorf("error setting up DNS: %w", err)
		}

		return nil
	}

	return firecracker.Handler{Name: "fcinit.SetupRootDrive", Fn: fn}
}
