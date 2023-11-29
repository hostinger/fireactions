package microvm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

// Manager is responsible for managing Micro VMs.
type Manager interface {
	// CreateMicroVM creates a new Micro VM.
	CreateMicroVM(ctx context.Context, spec Spec) (*MicroVM, error)

	// DeleteMicroVM deletes an existing Micro VM.
	DeleteMicroVM(ctx context.Context, id string) error

	// StartMicroVM starts an existing Micro VM.
	StartMicroVM(ctx context.Context, id string) error

	// GetMicroVMStatus returns the status of an existing Micro VM.
	GetMicroVMStatus(ctx context.Context, id string) (*Status, error)

	// GetMicroVM returns an existing Micro VM.
	GetMicroVM(ctx context.Context, id string) (*MicroVM, error)

	// StopMicroVM stops an existing Micro VM.
	StopMicroVM(ctx context.Context, id string) error

	// ListMicroVMs returns a list of all Micro VMs.
	ListMicroVMs(ctx context.Context) ([]*MicroVM, error)
}

type managerImpl struct {
	driver   Driver
	microvms map[string]*MicroVM
	l        sync.RWMutex
	logger   *zerolog.Logger
	wg       sync.WaitGroup
}

// Opt is a functional option for the Manager.
type Opt func(*managerImpl)

// WithLogger sets the logger for the Manager.
func WithLogger(logger *zerolog.Logger) Opt {
	f := func(m *managerImpl) {
		m.logger = logger
	}

	return f
}

// NewInMemoryManager creates a new in-memory Manager.
func NewInMemoryManager(driver Driver, opts ...Opt) Manager {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	m := &managerImpl{
		driver:   driver,
		microvms: make(map[string]*MicroVM),
		l:        sync.RWMutex{},
		logger:   &logger,
		wg:       sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// CreateMicroVM creates a new Micro VM. If the Micro VM with the same ID
// already exists, it returns the existing Micro VM.
func (m *managerImpl) CreateMicroVM(ctx context.Context, spec Spec) (*MicroVM, error) {
	microvm, ok := m.getMicroVM(spec.ID)
	if ok {
		return microvm, nil
	}

	microvm = &MicroVM{
		Spec:   spec,
		Status: Status{State: MicroVMStateStopped, Address: ""},
	}

	m.setMicroVM(microvm)
	return microvm, nil
}

// DeleteMicroVM deletes an existing Micro VM. If the Micro VM does not exist,
// it returns an error of type ErrNotFound.
func (m *managerImpl) DeleteMicroVM(ctx context.Context, id string) error {
	_, ok := m.getMicroVM(id)
	if !ok {
		return ErrNotFound
	}

	err := m.StopMicroVM(ctx, id)
	if err != nil {
		return err
	}

	m.deleteMicroVM(id)
	return nil
}

// StartMicroVM starts an existing Micro VM. If the Micro VM is already running,
// it does nothing. If the Micro VM does not exist, it returns an error of type
// ErrNotFound.
func (m *managerImpl) StartMicroVM(ctx context.Context, id string) error {
	microvm, ok := m.getMicroVM(id)
	if !ok {
		return ErrNotFound
	}

	if microvm.Status.State == MicroVMStateRunning {
		return nil
	}

	err := m.driver.StartMicroVM(ctx, microvm)
	if err != nil {
		return fmt.Errorf("driver: %w", err)
	}

	go func() {
		err := m.driver.WaitMicroVM(ctx, microvm.Spec.ID)
		if err != nil && err != context.Canceled && !strings.Contains(err.Error(), "signal: terminated") {
			m.logger.Error().Err(err).Msgf("unexpected Micro VM exit: %s", microvm.Spec.ID)
		}

		microvm.Status.State = MicroVMStateStopped
	}()

	return nil
}

// GetMicroVMStatus returns the status of an existing Micro VM. If the Micro VM
// does not exist, it returns an error of type ErrNotFound.
func (m *managerImpl) GetMicroVMStatus(ctx context.Context, id string) (*Status, error) {
	microvm, ok := m.getMicroVM(id)
	if !ok {
		return nil, ErrNotFound
	}

	return &microvm.Status, nil
}

// StopMicroVM stops an existing Micro VM.
func (m *managerImpl) StopMicroVM(ctx context.Context, id string) error {
	microvm, ok := m.getMicroVM(id)
	if !ok {
		return ErrNotFound
	}

	err := m.driver.StopMicroVM(ctx, microvm.Spec.ID)
	if err != nil {
		return fmt.Errorf("driver: %w", err)
	}

	return nil
}

// GetMicroVM returns an existing Micro VM. If the Micro VM does not exist, it
// returns an error of type ErrNotFound.
func (m *managerImpl) GetMicroVM(ctx context.Context, id string) (*MicroVM, error) {
	microvm, ok := m.getMicroVM(id)
	if !ok {
		return nil, ErrNotFound
	}

	return microvm, nil
}

// ListMicroVMs returns a list of all Micro VMs.
func (m *managerImpl) ListMicroVMs(ctx context.Context) ([]*MicroVM, error) {
	m.l.RLock()
	defer m.l.RUnlock()

	microvms := make([]*MicroVM, 0, len(m.microvms))
	for _, microvm := range m.microvms {
		microvms = append(microvms, microvm)
	}

	return microvms, nil
}

func (m *managerImpl) getMicroVM(id string) (*MicroVM, bool) {
	m.l.RLock()
	defer m.l.RUnlock()

	microvm, ok := m.microvms[id]
	if !ok {
		return nil, false
	}

	return microvm, true
}

func (m *managerImpl) setMicroVM(microvm *MicroVM) {
	m.l.Lock()
	defer m.l.Unlock()

	m.microvms[microvm.Spec.ID] = microvm
}

func (m *managerImpl) deleteMicroVM(id string) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.microvms, id)
}
