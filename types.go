package fireactions

import (
	"fmt"
	"time"
)

// Runner represents a GitHub Actions self-hosted runner instance.
type Runner struct {
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	NodeID          *string                     `json:"node_id"`
	ImagePullPolicy RunnerImagePullPolicy       `json:"image_pull_policy"`
	Image           string                      `json:"image"`
	Affinity        []*RunnerAffinityExpression `json:"affinity"`
	Status          RunnerStatus                `json:"status"`
	Organisation    string                      `json:"organisation"`
	Labels          []string                    `json:"labels"`
	Resources       RunnerResources             `json:"resources"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	DeletedAt       *time.Time                  `json:"deleted_at"`
}

type RunnerAffinityExpression struct {
	Key      string
	Operator string
	Values   []string
}

func (r *RunnerAffinityExpression) String() string {
	return fmt.Sprintf("%s %s %v", r.Key, r.Operator, r.Values)
}

type RunnerImagePullPolicy string

const (
	RunnerImagePullPolicyAlways       RunnerImagePullPolicy = "Always"
	RunnerImagePullPolicyIfNotPresent RunnerImagePullPolicy = "IfNotPresent"
	RunnerImagePullPolicyNever        RunnerImagePullPolicy = "Never"
)

type RunnerPhase string

const (
	RunnerPhasePending   RunnerPhase = "Pending"
	RunnerPhaseIdle      RunnerPhase = "Idle"
	RunnerPhaseActive    RunnerPhase = "Active"
	RunnerPhaseCompleted RunnerPhase = "Completed"
)

type RunnerStatus struct {
	Phase RunnerPhase
}

// RunnerFilterFunc is a function that filters a Runner. It returns true if the
// Runner should be included in the result set.
type RunnerFilterFunc func(*Runner) bool

// RunnerResources represents the resources required by a Runner. It is used to
// calculate the amount of resources that a Node needs to have available in
// order to run the Runner.
type RunnerResources struct {
	VCPUs       int64
	MemoryBytes int64
}

// Node represents a bare metal server that can run GitHub runners in
// Firecracker virtual machines.
type Node struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Labels            map[string]string `json:"labels"`
	Status            NodeStatus        `json:"status"`
	CPU               NodeResource      `json:"cpu"`
	RAM               NodeResource      `json:"ram"`
	HeartbeatInterval time.Duration     `json:"heartbeat_interval"`
	LastHeartbeat     time.Time         `json:"last_heartbeat"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// NodeStatus represents the status of a Node.
type NodeStatus string

const (
	// NodeStatusCordoned means that the Node is not available for scheduling.
	NodeStatusCordoned NodeStatus = "Cordoned"

	// NodeStatusReady means that the Node is available for scheduling.
	NodeStatusReady NodeStatus = "Ready"
)

// NodeFilterFunc is a function that filters a Node. It returns true if the
// Node should be included in the result set.
type NodeFilterFunc func(*Node) bool

// NodeResource represents a resource on a Node, e.g. CPU or RAM.
type NodeResource struct {
	Allocated       int64   `json:"allocated"`
	Capacity        int64   `json:"capacity"`
	OvercommitRatio float64 `json:"overcommit_ratio"`
}

// Reserve reserves the given amount of the resource.
func (r *NodeResource) Reserve(amount int64) {
	r.Allocated += amount
}

// Release releases the given amount of the resource.
func (r *NodeResource) Release(amount int64) {
	r.Allocated -= amount
}

// Available returns the amount of available resource.
func (r *NodeResource) Available() int64 {
	return r.Capacity - r.Allocated
}

// String returns a string representation of the resource.
func (r *NodeResource) String() string {
	return fmt.Sprintf("%d/%d", r.Allocated, r.Capacity)
}

// IsAvailable returns true if the given amount of resource is available. It
// takes the overcommit ratio into account.
func (r *NodeResource) IsAvailable(amount int64) bool {
	return float64(r.Allocated+amount) <= float64(r.Capacity)*r.OvercommitRatio
}

// IsFull returns true if the resource is fully allocated.
func (r *NodeResource) IsFull() bool {
	return r.Allocated >= r.Capacity
}
