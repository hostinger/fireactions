package cache

import (
	"sync"
	"time"

	"github.com/hostinger/fireactions"
)

// Cache is responsible for caching queued runners.
type Cache interface {
	// Get returns a queued runner.
	Get(key string) (*queuedRunner, bool)
	// Set sets a queued runner.
	Set(key string, runner *fireactions.Runner)
	// Delete deletes a queued runner.
	Delete(key string)
	// IncAttempts increments the number of attempts for a queued runner.
	IncAttempts(key string)
	// GetAttempts returns the number of attempts for a queued runner.
	GetAttempts(key string) int
}

type queuedRunnersCache struct {
	runners map[string]*queuedRunner
	mu      sync.Mutex
}

// NewCache creates a new Cache.
func NewCache() *queuedRunnersCache {
	q := &queuedRunnersCache{
		runners: make(map[string]*queuedRunner),
		mu:      sync.Mutex{},
	}

	return q
}

type queuedRunner struct {
	runner *fireactions.Runner

	EnqueueTime time.Time
	Attempts    int
}

// Get returns a queued runner.
func (q *queuedRunnersCache) Get(key string) (*queuedRunner, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	runner, ok := q.runners[key]
	return runner, ok
}

// Set sets a queued runner.
func (q *queuedRunnersCache) Set(key string, runner *fireactions.Runner) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.runners[key] = &queuedRunner{runner: runner, Attempts: 0, EnqueueTime: time.Now()}
}

// Delete deletes a queued runner.
func (q *queuedRunnersCache) Delete(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.runners, key)
}

// IncAttempts increments the number of attempts for a queued runner.
func (q *queuedRunnersCache) IncAttempts(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	runner, ok := q.runners[key]
	if !ok {
		return
	}

	if runner.Attempts < 60 {
		runner.Attempts++
	}
}

// GetAttempts returns the number of attempts for a queued runner.
func (q *queuedRunnersCache) GetAttempts(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.runners[key].Attempts
}
