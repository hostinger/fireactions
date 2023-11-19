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
	"github.com/hostinger/fireactions/client/microvm"
	"github.com/sirupsen/logrus"
)

// DriverConfig is the configuration for the Firecracker driver.
type DriverConfig struct {
	BinaryPath      string
	SocketPath      string
	KernelImagePath string
	KernelArgs      string
	CNIConfDir      string
	CNIBinDirs      []string
}

// Driver is the Firecracker driver.
type Driver struct {
	microVMs map[string]*microvm.MicroVM
	machines map[string]*firecracker.Machine
	config   *DriverConfig
	l        sync.RWMutex
}

// NewDriver creates a new Firecracker driver.
func NewDriver(config *DriverConfig) *Driver {
	m := &Driver{
		microVMs: make(map[string]*microvm.MicroVM),
		machines: make(map[string]*firecracker.Machine),
		config:   config,
		l:        sync.RWMutex{},
	}

	return m
}

// CreateVM creates a new Firecracker VM.
func (c *Driver) CreateVM(ctx context.Context, microvm *microvm.MicroVM) error {
	c.l.Lock()
	defer c.l.Unlock()

	_, ok := c.microVMs[microvm.ID]
	if ok {
		return nil
	}

	c.microVMs[microvm.ID] = microvm
	return nil
}

// StartVM starts a Firecracker VM.
func (c *Driver) StartVM(ctx context.Context, id string) error {
	c.l.Lock()
	defer c.l.Unlock()

	if !c.exists(ctx, id) {
		return microvm.ErrNotFound
	}

	_, ok := c.machines[id]
	if ok {
		return nil
	}

	fcMachine, err := c.createFirecrackerMachine(c.microVMs[id])
	if err != nil {
		return fmt.Errorf("creating Firecracker VM: %w", err)
	}

	// (konradasb): Socket might exist if the MicroVM was not properly stopped. Remove it to avoid errors.
	err = os.Remove(fcMachine.Cfg.SocketPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing socket file: %w", err)
	}

	if err := fcMachine.Start(context.Background()); err != nil {
		return fmt.Errorf("starting Firecracker VM: %w", err)
	}
	c.machines[id] = fcMachine

	ip := fcMachine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String()
	c.microVMs[id].Status = microvm.MicroVMStatus{State: microvm.MicroVMStateRunning, IP: ip}

	go c.watchFirecrackerMachine(fcMachine)
	return nil
}

// DeleteVM deletes a Firecracker VM.
func (c *Driver) DeleteVM(ctx context.Context, id string) error {
	c.l.Lock()
	defer c.l.Unlock()

	if !c.exists(ctx, id) {
		return microvm.ErrNotFound
	}

	machine, ok := c.machines[id]
	if !ok {
		return nil
	}

	if err := machine.StopVMM(); err != nil {
		return fmt.Errorf("stopping Firecracker VM: %w", err)
	}

	delete(c.microVMs, id)
	return nil
}

// StopVM stops a Firecracker VM.
func (c *Driver) StopVM(ctx context.Context, id string) error {
	c.l.Lock()
	defer c.l.Unlock()

	if !c.exists(ctx, id) {
		return microvm.ErrNotFound
	}

	machine, ok := c.machines[id]
	if !ok {
		return nil
	}

	if err := machine.StopVMM(); err != nil {
		return fmt.Errorf("stopping Firecracker VM: %w", err)
	}

	c.microVMs[id].Status = microvm.MicroVMStatus{State: microvm.MicroVMStateStopped, IP: ""}
	return nil
}

// WaitVM waits for a Firecracker VM to stop.
func (c *Driver) WaitVM(ctx context.Context, id string) error {
	c.l.Lock()
	defer c.l.Unlock()

	if !c.exists(ctx, id) {
		return microvm.ErrNotFound
	}

	machine, ok := c.machines[id]
	if !ok {
		return nil
	}

	return machine.Wait(ctx)
}

// ListVMs lists all Firecracker VMs.
func (c *Driver) ListVMs(ctx context.Context) ([]*microvm.MicroVM, error) {
	c.l.RLock()
	defer c.l.RUnlock()

	microVMs := []*microvm.MicroVM{}
	for _, m := range c.microVMs {
		microVMs = append(microVMs, m)
	}

	return microVMs, nil
}

// GetVM gets a Firecracker VM.
func (c *Driver) GetVM(ctx context.Context, id string) (*microvm.MicroVM, error) {
	c.l.RLock()
	defer c.l.RUnlock()

	m, ok := c.microVMs[id]
	if !ok {
		return nil, microvm.ErrNotFound
	}

	return m, nil
}

func (c *Driver) exists(ctx context.Context, id string) bool {
	_, ok := c.microVMs[id]
	return ok
}

func (c *Driver) createFirecrackerMachine(microvm *microvm.MicroVM) (*firecracker.Machine, error) {
	config := newFirecrackerConfigFromMicroVM(c.config, microvm)

	fcMachineCmd := firecracker.VMCommandBuilder{}.
		WithBin(c.config.BinaryPath).
		WithSocketPath(fmt.Sprintf(c.config.SocketPath, microvm.ID)).
		WithStdout(io.Discard).
		WithStderr(io.Discard).
		Build(context.Background())

	logger := logrus.New()
	logger.SetOutput(io.Discard)

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

	m, ok := c.microVMs[machine.Cfg.VMID]
	if !ok {
		return
	}

	m.Status = microvm.MicroVMStatus{State: microvm.MicroVMStateStopped, IP: ""}
	delete(c.machines, m.ID)
}

func newFirecrackerConfigFromMicroVM(config *DriverConfig, microvm *microvm.MicroVM) *firecracker.Config {
	fc := firecracker.Config{
		VMID:              microvm.ID,
		SocketPath:        fmt.Sprintf(config.SocketPath, microvm.ID),
		KernelImagePath:   config.KernelImagePath,
		KernelArgs:        config.KernelArgs,
		Drives:            []models.Drive{},
		NetworkInterfaces: []firecracker.NetworkInterface{},
		MmdsAddress:       net.IPv4(169, 254, 169, 254),
		MmdsVersion:       firecracker.MMDSv2,
		ForwardSignals:    []os.Signal{os.Interrupt},
		LogLevel:          "debug",
	}

	memSizeMib := microvm.Spec.MemoryBytes / 1024 / 1024
	fc.MachineCfg = models.MachineConfiguration{VcpuCount: &microvm.Spec.VCPU, MemSizeMib: &memSizeMib}

	for _, networkInterface := range microvm.Spec.NetworkInterfaces {
		n := firecracker.NetworkInterface{AllowMMDS: true, CNIConfiguration: &firecracker.CNIConfiguration{
			NetworkName: networkInterface.NetworkName,
			IfName:      networkInterface.IfName,
			VMIfName:    networkInterface.VMIfName,
			ConfDir:     config.CNIConfDir,
			BinPath:     config.CNIBinDirs,
		}}

		if networkInterface.ConfDir == "" {
			n.CNIConfiguration.ConfDir = config.CNIConfDir
		}

		if networkInterface.BinPath == "" {
			n.CNIConfiguration.BinPath = config.CNIBinDirs
		}

		fc.NetworkInterfaces = append(fc.NetworkInterfaces, n)
	}

	for _, drive := range microvm.Spec.Drives {
		d := models.Drive{DriveID: &drive.ID, PathOnHost: &drive.PathOnHost, IsReadOnly: &drive.IsReadOnly, IsRootDevice: &drive.IsRoot}
		fc.Drives = append(fc.Drives, d)
	}

	return &fc
}
