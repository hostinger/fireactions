package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/client/dmsetup"
	"github.com/hostinger/fireactions/client/losetup"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
)

// RunnerConfig is the configuration for a Runner.
type RunnerConfig struct {
	ID           uuid.UUID
	Name         string
	Organisation string
	Labels       string
	Token        string
	VCPUs        int64
	MemorySizeMB int64
	DiskSizeGB   int64
	Image        string
}

// Runner is a representation of a Firecracker VM with GitHub Actions self-hosted runner.
type Runner struct {
	Config *RunnerConfig

	machine         *firecracker.Machine
	machineMetadata map[string]interface{}
	dmsetup         *dmsetup.Client
	losetup         *losetup.Client
	cleanupFuncs    []func() error
	cleanupOnce     sync.Once
	isCleanedUp     bool
	log             *zerolog.Logger
}

// RunnerOpt is a function that modifies a Runner.
type RunnerOpt func(*Runner)

// NewRunner creates a new Runner.
func NewRunner(log *zerolog.Logger, config *RunnerConfig, opts ...RunnerOpt) (*Runner, error) {
	r := &Runner{
		Config:          config,
		dmsetup:         dmsetup.DefaultClient(),
		losetup:         losetup.DefaultClient(),
		cleanupFuncs:    make([]func() error, 0),
		cleanupOnce:     sync.Once{},
		isCleanedUp:     false,
		machineMetadata: make(map[string]interface{}),
		log:             log,
	}

	for _, opt := range opts {
		opt(r)
	}

	fc := firecracker.Config{
		VMID:            r.Config.ID.String(),
		MachineCfg:      models.MachineConfiguration{MemSizeMib: &r.Config.MemorySizeMB, VcpuCount: &r.Config.VCPUs},
		SocketPath:      filepath.Join("/var/run", fmt.Sprintf("fireactions-%s.socket", r.Config.ID.String())),
		KernelImagePath: "/var/lib/fireactions/vmlinux.bin",
		KernelArgs:      "ro console=ttyS0 noapic reboot=k panic=1 pci=off nomodules random.trust_cpu=on",
		MmdsVersion:     firecracker.MMDSv2,
		MmdsAddress:     net.IPv4(169, 254, 169, 254),
		LogPath:         filepath.Join("/var/log/fireactions", fmt.Sprintf("%s.log", r.Config.ID.String())),
		LogLevel:        "debug",
	}
	fc.Drives = append(fc.Drives, models.Drive{
		DriveID:      firecracker.String("1"),
		PathOnHost:   firecracker.String(fmt.Sprintf("/dev/mapper/fireactions-%s", r.Config.ID.String())),
		IsRootDevice: firecracker.Bool(true),
		IsReadOnly:   firecracker.Bool(false),
	})
	fc.NetworkInterfaces = append(fc.NetworkInterfaces, firecracker.NetworkInterface{
		AllowMMDS:        true,
		CNIConfiguration: &firecracker.CNIConfiguration{NetworkName: "fireactions", IfName: "veth0"},
	})

	logf, err := os.OpenFile(fc.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening log file %s: %v", fc.LogPath, err)
	}

	r.addCleanupFunc(func() error {
		return logf.Close()
	})

	cmd := firecracker.VMCommandBuilder{}.
		WithBin("firecracker").
		WithSocketPath(fc.SocketPath).
		WithStdout(logf).
		WithStderr(logf).
		Build(context.Background())

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	m, err := firecracker.NewMachine(context.Background(), fc, firecracker.WithProcessRunner(cmd), firecracker.WithLogger(logrus.NewEntry(logger)))
	if err != nil {
		return nil, fmt.Errorf("error creating Firecracker machine: %v", err)
	}

	r.machine = m
	r.machineMetadata = map[string]interface{}{"latest": map[string]interface{}{"meta-data": map[string]interface{}{
		"runner-id":     r.Config.ID.String(),
		"runner-name":   r.Config.Name,
		"runner-url":    fmt.Sprintf("https://github.com/%s", r.Config.Organisation),
		"runner-labels": r.Config.Labels,
		"runner-token":  r.Config.Token,
	}}}

	return r, nil
}

func (r *Runner) Start() error {
	ctx := context.Background()

	err := r.setup(ctx)
	if err != nil {
		return fmt.Errorf("error setting up runner: %v", err)
	}

	err = r.machine.Start(context.Background())
	if err != nil {
		return fmt.Errorf("error starting machine: %v", err)
	}

	err = r.machine.SetMetadata(ctx, r.machineMetadata)
	if err != nil {
		return fmt.Errorf("error setting metadata: %v", err)
	}

	return nil
}

// Wait waits for Runner to exit. It also performs cleanup operations after the VM exits.
func (r *Runner) Wait() error {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		r.machine.Wait(ctx)
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		err := r.machine.StopVMM()
		if err != nil {
			r.log.Error().Err(err).Msg("error stopping VMM")
		}
	case <-stopped:
	}

	r.runCleanupFuncs()
	return nil
}

func (r *Runner) setup(ctx context.Context) error {
	loop1, err := r.losetup.Attach(ctx, filepath.Join("/var/lib/fireactions/images", r.Config.Image, "rootfs.ext4"))
	if err != nil {
		return fmt.Errorf("error attaching loop device: %v", err)
	}

	r.addCleanupFunc(func() error {
		err := r.losetup.Detach(ctx, loop1)
		if err != nil && strings.Contains(err.Error(), "No such device") {
			return nil
		}

		return err
	})

	loop1f, err := os.OpenFile(loop1, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error opening loop device: %v", err)
	}
	defer loop1f.Close()

	var size int64
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, loop1f.Fd(), uintptr(0x1260), uintptr(unsafe.Pointer(&size)))
	if errno != 0 {
		return errno
	}

	p := filepath.Join("/var/lib/fireactions/images", fmt.Sprintf("%s.ext4", r.Config.ID))
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("error creating image: %v", err)
	}
	defer f.Close()

	err = f.Truncate(r.Config.DiskSizeGB * 1024 * 1024 * 1024)
	if err != nil {
		return fmt.Errorf("error truncating image: %v", err)
	}

	r.addCleanupFunc(func() error {
		return os.Remove(p)
	})

	loop2, err := r.losetup.Attach(ctx, p)
	if err != nil {
		return fmt.Errorf("error attaching loop device: %v", err)
	}

	r.addCleanupFunc(func() error {
		err := r.losetup.Detach(ctx, loop2)
		if err != nil && strings.Contains(err.Error(), "No such device") {
			return nil
		}

		return err
	})

	err = r.dmsetup.Create(ctx, fmt.Sprintf("fireactions-%s", r.Config.ID), fmt.Sprintf("0 %d snapshot %s %s p 8", size, loop1, loop2))
	if err != nil && err.Error() != fmt.Sprintf("device %s already exists", r.Config.ID) {
		return fmt.Errorf("error creating device: %v", err)
	}

	r.addCleanupFunc(func() error {
		return r.dmsetup.Remove(ctx, fmt.Sprintf("fireactions-%s", r.Config.ID))
	})

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("fireactions-%s", r.Config.ID))
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = syscall.Mount(fmt.Sprintf("/dev/mapper/fireactions-%s", r.Config.ID.String()), tmpDir, "ext4", 0, "")
	if err != nil {
		return fmt.Errorf("error mounting disk: %v", err)
	}
	defer syscall.Unmount(tmpDir, 0)

	err = r.setupHostname(tmpDir)
	if err != nil {
		return fmt.Errorf("error setting up hostname: %v", err)
	}

	err = r.setupDNS(tmpDir)
	if err != nil {
		return fmt.Errorf("error setting up DNS: %v", err)
	}

	err = r.setupSSH()
	if err != nil {
		return fmt.Errorf("error setting up SSH: %v", err)
	}

	return nil
}

func (r *Runner) setupHostname(rootPath string) error {
	hostname := r.Config.Name

	err := os.WriteFile(filepath.Join(rootPath, "etc", "hostname"), []byte(fmt.Sprintf("%s\n", hostname)), 0644)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", filepath.Join(rootPath, "etc", "hostname"), err)
	}

	return nil
}

func (r *Runner) setupDNS(rootPath string) error {
	resolvConfPath := filepath.Join(rootPath, "etc", "resolv.conf")

	data, err := os.ReadFile(resolvConfPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", resolvConfPath, err)
	}

	if len(data) != 0 {
		return nil
	}

	os.Remove(resolvConfPath)
	err = os.Symlink("../proc/net/pnp", resolvConfPath)
	if err != nil {
		return fmt.Errorf("error creating symlink /etc/resolv.conf -> ../proc/net/pnp: %v", err)
	}

	return nil
}

func (r *Runner) setupSSH() error {
	return nil
}

func (r *Runner) addCleanupFunc(f func() error) {
	r.cleanupFuncs = append(r.cleanupFuncs, f)
}

func (r *Runner) runCleanupFuncs() {
	r.cleanupOnce.Do(func() {
		var err *multierror.Error

		for i := len(r.cleanupFuncs) - 1; i >= 0; i-- {
			err = multierror.Append(err, r.cleanupFuncs[i]())
		}

		if err.ErrorOrNil() != nil {
			r.log.Error().Err(err).Msg("error cleaning up")
			return
		}

		r.isCleanedUp = true
	})
}
