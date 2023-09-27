package firecracker

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/client/microvm"
	"github.com/sirupsen/logrus"
)

// DriverConfig is the configuration for the Firecracker driver.
type DriverConfig struct {
	BinaryPath      string `mapstructure:"binary_path"`
	SocketPath      string `mapstructure:"socket_path"`
	KernelImagePath string `mapstructure:"kernel_image_path"`
	KernelArgs      string `mapstructure:"kernel_args"`
	LogPath         string `mapstructure:"log_path"`
	LogLevel        string `mapstructure:"log_level"`
}

// NewDriverConfig creates a new DriverConfig with default values.
func NewDriverConfig() *DriverConfig {
	c := &DriverConfig{
		BinaryPath:      "firecracker",
		SocketPath:      "/var/run/fireactions-%s.sock",
		KernelImagePath: "vmlinux.bin",
		KernelArgs:      "console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw",
		LogPath:         "/var/log/fireactions-%s.log",
		LogLevel:        "info",
	}

	return c
}

// Validate validates the DriverConfig.
func (c *DriverConfig) Validate() error {
	var errs error

	if c.BinaryPath == "" {
		errs = multierror.Append(errs, fmt.Errorf("binary_path is required"))
	}

	if c.SocketPath == "" {
		errs = multierror.Append(errs, fmt.Errorf("socket_path is required"))
	}

	if c.KernelImagePath == "" {
		errs = multierror.Append(errs, fmt.Errorf("kernel_image_path is required"))
	}

	if c.KernelArgs == "" {
		errs = multierror.Append(errs, fmt.Errorf("kernel_args is required"))
	}

	if c.LogPath == "" {
		errs = multierror.Append(errs, fmt.Errorf("log_path is required"))
	}

	if c.LogLevel == "" {
		errs = multierror.Append(errs, fmt.Errorf("log_level is required"))
	}

	return errs
}

// Driver is the Firecracker driver.
type Driver struct {
	config   *DriverConfig
	machines map[string]*firecracker.Machine
	l        sync.RWMutex
}

// NewDriver creates a new Firecracker driver.
func NewDriver(config *DriverConfig) *Driver {
	m := &Driver{
		config:   config,
		machines: make(map[string]*firecracker.Machine),
		l:        sync.RWMutex{},
	}

	return m
}

// StartMicroVM starts a Firecracker VM.
func (c *Driver) StartMicroVM(ctx context.Context, m *microvm.MicroVM) error {
	_, ok := c.getMachine(m.Spec.ID)
	if ok {
		return nil
	}

	fcMachine, err := c.createFirecrackerMachine(m)
	if err != nil {
		return fmt.Errorf("creating Firecracker VM: %w", err)
	}

	// (konradasb): Socket might exist if the MicroVM was not properly stopped. Remove it to avoid errors.
	err = os.Remove(fcMachine.Cfg.SocketPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing socket file: %w", err)
	}

	if err := fcMachine.Start(context.Background()); err != nil {
		return err
	}

	go c.watchFirecrackerMachine(fcMachine)
	c.setMachine(fcMachine)

	ip := fcMachine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String()
	m.Status = microvm.Status{State: microvm.MicroVMStateRunning, Address: ip}

	return nil
}

// StopMicroVM stops a Firecracker VM.
func (c *Driver) StopMicroVM(ctx context.Context, id string) error {
	machine, ok := c.getMachine(id)
	if !ok {
		return nil
	}

	if err := machine.StopVMM(); err != nil {
		return err
	}

	err := machine.Wait(ctx)
	if err != nil && err == context.DeadlineExceeded {
		return err
	}

	c.delMachine(id)
	return nil
}

// WaitMicroVM waits for a Firecracker VM to stop.
func (c *Driver) WaitMicroVM(ctx context.Context, id string) error {
	machine, ok := c.getMachine(id)
	if !ok {
		return nil
	}

	return machine.Wait(ctx)
}

func (c *Driver) getMachine(id string) (*firecracker.Machine, bool) {
	c.l.RLock()
	defer c.l.RUnlock()

	m, ok := c.machines[id]
	return m, ok
}

func (c *Driver) setMachine(m *firecracker.Machine) {
	c.l.Lock()
	defer c.l.Unlock()

	c.machines[m.Cfg.VMID] = m
}

func (c *Driver) delMachine(id string) {
	c.l.Lock()
	defer c.l.Unlock()

	delete(c.machines, id)
}

func (c *Driver) createFirecrackerMachine(microvm *microvm.MicroVM) (*firecracker.Machine, error) {
	config := newFirecrackerConfigFromMicroVM(c.config, microvm)

	fcMachineCmd := firecracker.VMCommandBuilder{}.
		WithBin(c.config.BinaryPath).
		WithSocketPath(fmt.Sprintf(c.config.SocketPath, microvm.Spec.ID)).
		WithStdout(io.Discard).
		WithStderr(io.Discard).
		Build(context.Background())

	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.DebugLevel)

	fcMachineOpts := []firecracker.Opt{firecracker.WithProcessRunner(fcMachineCmd), firecracker.WithLogger(logrus.NewEntry(logger))}
	fcMachine, err := firecracker.NewMachine(context.Background(), *config, fcMachineOpts...)
	if err != nil {
		return nil, err
	}

	metadata := map[string]interface{}{"latest": map[string]interface{}{"meta-data": map[string]interface{}{}}}
	metadata["latest"].(map[string]interface{})["meta-data"] = microvm.Spec.Metadata

	fcMachine.Handlers.FcInit = fcMachine.Handlers.FcInit.Append(firecracker.NewSetMetadataHandler(metadata))
	return fcMachine, nil
}

func (c *Driver) watchFirecrackerMachine(machine *firecracker.Machine) {
	machine.Wait(context.Background())

	c.l.Lock()
	defer c.l.Unlock()

	m, ok := c.machines[machine.Cfg.VMID]
	if !ok {
		return
	}

	delete(c.machines, m.Cfg.VMID)
}

func newFirecrackerConfigFromMicroVM(config *DriverConfig, microvm *microvm.MicroVM) *firecracker.Config {
	fc := firecracker.Config{
		VMID:              microvm.Spec.ID,
		SocketPath:        fmt.Sprintf(config.SocketPath, microvm.Spec.ID),
		KernelImagePath:   config.KernelImagePath,
		KernelArgs:        config.KernelArgs,
		Drives:            []models.Drive{},
		NetworkInterfaces: []firecracker.NetworkInterface{},
		MmdsAddress:       net.IPv4(169, 254, 169, 254),
		MmdsVersion:       firecracker.MMDSv2,
		ForwardSignals:    []os.Signal{},
		LogLevel:          config.LogLevel,
		LogPath:           fmt.Sprintf(config.LogPath, microvm.Spec.ID),
	}

	memSizeMib := microvm.Spec.MemoryBytes / 1024 / 1024
	fc.MachineCfg = models.MachineConfiguration{VcpuCount: &microvm.Spec.VCPU, MemSizeMib: &memSizeMib}

	for _, networkInterface := range microvm.Spec.NetworkInterfaces {
		n := firecracker.NetworkInterface{AllowMMDS: true, CNIConfiguration: &firecracker.CNIConfiguration{
			NetworkName: networkInterface.NetworkName,
			IfName:      networkInterface.IfName,
			VMIfName:    networkInterface.VMIfName,
			ConfDir:     "/etc/cni/net.d",
			BinPath:     []string{"/opt/cni/bin"},
		}}

		fc.NetworkInterfaces = append(fc.NetworkInterfaces, n)
	}

	for _, drive := range microvm.Spec.Drives {
		d := models.Drive{DriveID: &drive.ID, PathOnHost: &drive.PathOnHost, IsReadOnly: &drive.IsReadOnly, IsRootDevice: &drive.IsRoot}
		fc.Drives = append(fc.Drives, d)
	}

	return &fc
}
