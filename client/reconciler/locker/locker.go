package locker

import (
	"context"
	"fmt"
	"sync"

	"github.com/hostinger/fireactions"
)

// Locker is responsible for locking runners.
type Locker interface {
	// Acquire acquires a lock for the runner.
	Acquire(ctx context.Context, runner *fireactions.Runner) error
	// Release releases a lock for the runner.
	Release(ctx context.Context, runner *fireactions.Runner) error
}

type memoryLocker struct {
	runners map[string]struct{}
	mu      sync.Mutex
}

// NewMemoryLocker creates a new in-memory Locker.
func NewMemoryLocker() *memoryLocker {
	l := &memoryLocker{
		runners: make(map[string]struct{}),
		mu:      sync.Mutex{},
	}

	return l
}

// Acquire acquires a lock for the runner.
func (m *memoryLocker) Acquire(ctx context.Context, runner *fireactions.Runner) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.runners[runner.ID]; ok {
		return fmt.Errorf("lock for runner %s is already acquired", runner.ID)
	}

	m.runners[runner.ID] = struct{}{}
	return nil
}

// Release releases a lock for the runner.
func (m *memoryLocker) Release(ctx context.Context, runner *fireactions.Runner) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.runners, runner.ID)
	return nil
}
