package fireactions

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrPoolNotFound is returned when a pool is not found
	ErrPoolNotFound = errors.New("pool not found")
)

// Pool represents a slice of Pool
type Pools []*Pool

// Pool represents a pool of GitHub runners
type Pool struct {
	Name            string     `json:"name"`
	Replicas        int        `json:"replicas"`
	CurrentReplicas int        `json:"current_replicas"`
	DesiredReplicas int        `json:"desired_replicas"`
	Organization    string     `json:"organization"`
	GroupID         int64      `json:"group_id"`
	Labels          []string   `json:"labels"`
	Image           string     `json:"image"`
	Status          PoolStatus `json:"status"`
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
	return []string{"Name", "Current", "Desired", "Organization", "Group ID", "Labels", "Image", "State"}
}

func (p *Pool) ColsMap() map[string]string {
	return map[string]string{
		"Name":         "Name",
		"Current":      "Current",
		"Desired":      "Desired",
		"Organization": "Organization",
		"Group ID":     "Group ID",
		"Labels":       "Labels",
		"Image":        "Image",
		"State":        "State",
	}
}

func (p *Pool) KV() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"Name":         p.Name,
			"Current":      p.CurrentReplicas,
			"Desired":      p.DesiredReplicas,
			"Organization": p.Organization,
			"Group ID":     p.GroupID,
			"Labels":       strings.Join(p.Labels, ", "),
			"Image":        p.Image,
			"State":        p.Status.State,
		},
	}
}

func (p Pools) Cols() []string {
	return []string{"Name", "Current", "Desired", "Organization", "Group ID", "Labels", "Image", "State"}
}

func (p Pools) ColsMap() map[string]string {
	return map[string]string{
		"Name":         "Name",
		"Current":      "Current",
		"Desired":      "Desired",
		"Organization": "Organization",
		"Group ID":     "Group ID",
		"Labels":       "Labels",
		"Image":        "Image",
		"State":        "State",
	}
}

func (p Pools) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(p))
	for _, pool := range p {
		kv = append(kv, map[string]interface{}{
			"Name":         pool.Name,
			"Current":      pool.CurrentReplicas,
			"Desired":      pool.DesiredReplicas,
			"Organization": pool.Organization,
			"Group ID":     pool.GroupID,
			"Labels":       strings.Join(pool.Labels, ", "),
			"Image":        pool.Image,
			"State":        pool.Status.State,
		})
	}

	return kv
}

// MicroVMs represents a slice of MicroVM
type MicroVMs []MicroVM

// MicroVM represents a Firecracker based virtual machine
type MicroVM struct {
	VMID      string    `json:"VMID"`
	Pool      string    `json:"Pool"`
	IPAddr    string    `json:"IPAddr"`
	CreatedAt time.Time `json:"CreatedAt"`
}

func (m MicroVMs) Cols() []string {
	return []string{"Pool", "VMID", "IP Address", "Created"}
}

func (m MicroVMs) ColsMap() map[string]string {
	return map[string]string{"Pool": "Pool", "VMID": "VMID", "IP Address": "IP Address", "Created": "Created"}
}

func (m MicroVMs) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(m))
	for _, vm := range m {
		kv = append(kv, map[string]interface{}{
			"Pool":       vm.Pool,
			"VMID":       vm.VMID,
			"IP Address": vm.IPAddr,
			"Created":    formatDuration(time.Since(vm.CreatedAt)),
		})
	}
	return kv
}

// formatDuration converts a duration to a human-readable format
func formatDuration(d time.Duration) string {
	// Handle negative durations (future timestamps)
	if d < 0 {
		d = -d
		if d < time.Second {
			return "in a moment"
		}
		if d < time.Minute {
			seconds := int(d.Seconds())
			if seconds == 1 {
				return "in 1 second"
			}
			return fmt.Sprintf("in %d seconds", seconds)
		}
		if d < time.Hour {
			minutes := int(d.Minutes())
			if minutes == 1 {
				return "in 1 minute"
			}
			return fmt.Sprintf("in %d minutes", minutes)
		}
		if d < 24*time.Hour {
			hours := int(d.Hours())
			if hours == 1 {
				return "in 1 hour"
			}
			return fmt.Sprintf("in %d hours", hours)
		}
		days := int(d.Hours() / 24)
		if days == 1 {
			return "in 1 day"
		}
		return fmt.Sprintf("in %d days", days)
	}

	if d < time.Second {
		return "just now"
	}
	if d < time.Minute {
		seconds := int(d.Seconds())
		if seconds == 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
