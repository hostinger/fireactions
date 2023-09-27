package microvm

import (
	"errors"
)

var (
	// ErrNotFound is returned when a MicroVM is not found.
	ErrNotFound = errors.New("not found")
)

// MicroVM represents a virtual machine.
type MicroVM struct {
	Spec   Spec
	Status Status
}

// Status is the status of a MicroVM.
type Status struct {
	Address string
	State   MicroVMState
}

// MicroVMSpec is the specification of a MicroVM.
type Spec struct {
	ID                string
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
