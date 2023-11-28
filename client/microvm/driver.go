package microvm

import "context"

// Driver is the interface that wraps the basic methods for managing MicroVMs.
type Driver interface {
	// StartMicroVM starts a MicroVM.
	StartMicroVM(ctx context.Context, microvm *MicroVM) error

	// StopMicroVM stops a MicroVM.
	StopMicroVM(ctx context.Context, id string) error

	// WaitMicroVM waits for a MicroVM to stop.
	WaitMicroVM(ctx context.Context, id string) error
}
