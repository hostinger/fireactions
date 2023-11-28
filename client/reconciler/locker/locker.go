package locker

import (
	"context"
	"fmt"
	"sync"
)

// Locker is responsible for locking runners.
type Locker interface {
	// Acquire acquires a lock for the runner.
	Acquire(ctx context.Context, runnerID string) error
	// Release releases a lock for the runner.
	Release(ctx context.Context, runnerID string) error
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

// Acquire acquires a lock for the runner ID.
func (m *memoryLocker) Acquire(ctx context.Context, runnerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.runners[runnerID]; ok {
		return fmt.Errorf("lock for runner %s is already acquired", runnerID)
	}

	m.runners[runnerID] = struct{}{}
	return nil
}

// Release releases a lock for the runner ID.
func (m *memoryLocker) Release(ctx context.Context, runnerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.runners, runnerID)
	return nil
}
