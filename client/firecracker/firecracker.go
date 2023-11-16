package firecracker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/sirupsen/logrus"
)

// MachineReadinessProbe is a function that checks if a virtual machine is ready
// to accept connections.
type MachineReadinessProbe func(ctx context.Context, address string) error

// Machine represents a Firecracker virtual machine.
type Machine struct {
	machine        *firecracker.Machine
	machineConfig  *Config
	stdout         io.Writer
	stderr         io.Writer
	readinessProbe MachineReadinessProbe
	exitCh         chan error
}

// Opt is a functional option for configuring a Machine.
type Opt func(*Machine)

// WithStdout sets the stdout writer for logs emitted by the virtual machine.
func WithStdout(w io.Writer) Opt {
	f := func(m *Machine) {
		m.stdout = w
	}

	return f
}

// WithStderr sets the stderr writer for logs emitted by the virtual machine.
func WithStderr(w io.Writer) Opt {
	f := func(m *Machine) {
		m.stderr = w
	}

	return f
}

// WithReadinessProbe sets the readiness probe for the virtual machine.
func WithReadinessProbe(readinessProbe MachineReadinessProbe) Opt {
	f := func(m *Machine) {
		m.readinessProbe = readinessProbe
	}

	return f
}

// NewMachine creates a new Machine.
func NewMachine(config *Config, opts ...Opt) *Machine {
	m := &Machine{
		machineConfig:  config,
		machine:        nil,
		readinessProbe: func(ctx context.Context, address string) error { return nil },
		stdout:         io.Discard,
		stderr:         io.Discard,
		exitCh:         make(chan error),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Start starts the virtual machine.
func (m *Machine) Start(ctx context.Context) error {
	if m.IsRunning() {
		return nil
	}

	machineCmd := firecracker.VMCommandBuilder{}.
		WithBin("firecracker").
		WithSocketPath(m.machineConfig.SocketPath).
		WithStdout(m.stdout).
		WithStderr(m.stderr).
		Build(context.TODO())

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	machine, err := firecracker.NewMachine(ctx, *m.machineConfig.Config, firecracker.WithProcessRunner(machineCmd), firecracker.WithLogger(logrus.NewEntry(logger)))
	if err != nil {
		return fmt.Errorf("firecracker: %w", err)
	}
	m.machine = machine
	m.machine.Handlers.FcInit = machine.Handlers.FcInit.Append(firecracker.NewSetMetadataHandler(m.machineConfig.Metadata))

	err = machine.Start(context.Background())
	if err != nil {
		return fmt.Errorf("firecracker: %w", err)
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}

		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		err := m.readinessProbe(ctx, m.machineConfig.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String())
		if err != nil {
			continue
		}

		break
	}

	go func() {
		m.exitCh <- m.machine.Wait(context.Background())
	}()

	return nil
}

// Stop stops the virtual machine.
func (m *Machine) Stop(ctx context.Context) error {
	if !m.IsRunning() {
		return nil
	}

	err := m.machine.StopVMM()
	if err != nil {
		return fmt.Errorf("firecracker: %w", err)
	}

	err = m.machine.Wait(ctx)
	if err != nil && err == context.DeadlineExceeded {
		return fmt.Errorf("firecracker: %w", err)
	}

	return nil
}

// IsRunning returns true if the virtual machine is running.
func (m *Machine) IsRunning() bool {
	if m.machine == nil {
		return false
	}

	_, err := m.machine.PID()
	if err != nil {
		return false
	}

	return true
}

// ExitCh returns a channel that will receive an error when the virtual machine exits.
func (m *Machine) ExitCh() <-chan error {
	return m.exitCh
}

// String returns a string representation of the Machine.
func (m *Machine) String() string {
	return m.ID()
}

// ID returns the ID of the Machine.
func (m *Machine) ID() string {
	return m.machineConfig.VMID
}

func (m *Machine) Config() *Config {
	return m.machineConfig
}
