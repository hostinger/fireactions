package microvm

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned when a MicroVM is not found.
	ErrNotFound = errors.New("not found")
)

// MicroVM represents a virtual machine.
type MicroVM struct {
	// ID is the unique identifier of the MicroVM.
	ID string

	// Spec is the specification of the MicroVM.
	Spec MicroVMSpec

	// Status is the current status of the MicroVM.
	Status MicroVMStatus
}

// MicroVMSpec is the specification of a MicroVM.
type MicroVMSpec struct {
	Name              string
	VCPU              int64
	MemoryBytes       int64
	Metadata          map[string]interface{}
	Drives            []Drive
	NetworkInterfaces []NetworkInterface
}

// Drive is a drive attached to a MicroVM.
type Drive struct {
	ID         string
	PathOnHost string
	IsReadOnly bool
	IsRoot     bool
}

// NetworkInterface is a network interface attached to a MicroVM.
type NetworkInterface struct {
	AllowMMDS   bool
	NetworkName string
	IfName      string
	VMIfName    string
	ConfDir     string
	BinPath     string
}

// MicroVMStatus is the status of a MicroVM.
type MicroVMStatus struct {
	IP    string
	State MicroVMState
}

// MicroVMState is the state of a MicroVM.
type MicroVMState string

const (
	// MicroVMStateUnknown is the state of a MicroVM when it is unknown.
	MicroVMStateUnknown MicroVMState = "Unknown"

	// MicroVMStateRunning is the state of a MicroVM when it is running.
	MicroVMStateRunning MicroVMState = "Running"

	// MicroVMStateStopped is the state of a MicroVM when it is stopped.
	MicroVMStateStopped MicroVMState = "Stopped"
)

// Driver is the interface that wraps the basic methods for managing MicroVMs.
type Driver interface {
	// CreateVM creates a new MicroVM.
	CreateVM(ctx context.Context, microVM *MicroVM) error

	// DeleteVM deletes a MicroVM.
	DeleteVM(ctx context.Context, id string) error

	// GetVM returns a MicroVM.
	GetVM(ctx context.Context, id string) (*MicroVM, error)

	// ListVMs returns a list of MicroVMs.
	ListVMs(ctx context.Context) ([]*MicroVM, error)

	// StartVM starts a MicroVM.
	StartVM(ctx context.Context, id string) error

	// StopVM stops a MicroVM.
	StopVM(ctx context.Context, id string) error

	// WaitVM waits for a MicroVM to stop.
	WaitVM(ctx context.Context, id string) error
}
