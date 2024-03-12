package fireactions

import (
	"errors"
)

var (
	// ErrPoolNotFound is returned when a pool is not found
	ErrPoolNotFound = errors.New("pool not found")
)

// Pool represents a slice of Pool
type Pools []*Pool

// Pool represents a pool of GitHub runners
type Pool struct {
	Name       string     `json:"name"`
	MaxRunners int        `json:"max_runners"`
	MinRunners int        `json:"min_runners"`
	CurRunners int        `json:"cur_runners"`
	Status     PoolStatus `json:"status"`
}

// PoolState represents the state of a pool
type PoolState string

// String returns the string representation of the pool state
func (p PoolState) String() string {
	return string(p)
}

const (
	// PoolStateActive represents the active state, meaning the pool is running
	PoolStateActive PoolState = "Active"

	// PoolStatePaused represents the paused state, meaning the pool is stopped
	PoolStatePaused PoolState = "Paused"
)

// PoolStatus represents the status of a pool
type PoolStatus struct {
	State   PoolState `json:"state"`
	Message string    `json:"message"`
}

func (p *Pool) Cols() []string {
	return []string{"Name", "Max Runners", "Min Runners", "Cur Runners", "State"}
}

func (p *Pool) ColsMap() map[string]string {
	return map[string]string{
		"Name":       "Name",
		"MaxRunners": "Max Runners",
		"MinRunners": "Min Runners",
		"CurRunners": "Cur Runners",
		"State":      "State",
	}
}

func (p *Pool) KV() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"Name":        p.Name,
			"Max Runners": p.MaxRunners,
			"Min Runners": p.MinRunners,
			"Cur Runners": p.CurRunners,
			"State":       p.Status.State,
		},
	}
}

func (p Pools) Cols() []string {
	return []string{"Name", "Max Runners", "Min Runners", "Cur Runners", "State"}
}

func (p Pools) ColsMap() map[string]string {
	return map[string]string{
		"Name":       "Name",
		"MaxRunners": "Max Runners",
		"MinRunners": "Min Runners",
		"CurRunners": "Cur Runners",
		"State":      "State",
	}
}

func (p Pools) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(p))
	for _, pool := range p {
		kv = append(kv, map[string]interface{}{
			"Name":        pool.Name,
			"Max Runners": pool.MaxRunners,
			"Min Runners": pool.MinRunners,
			"Cur Runners": pool.CurRunners,
			"State":       pool.Status.State,
		})
	}

	return kv
}
