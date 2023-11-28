package fireactions

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Runner represents a GitHub Actions self-hosted runner instance.
type Runner struct {
	ID              string                 `json:"id"`
	NodeID          *string                `json:"node_id"`
	Name            string                 `json:"name"`
	Organisation    string                 `json:"organisation"`
	Image           string                 `json:"image"`
	ImagePullPolicy RunnerImagePullPolicy  `json:"image_pull_policy"`
	Metadata        map[string]interface{} `json:"metadata"`
	Affinity        []*RunnerAffinityRule  `json:"affinity"`
	Resources       RunnerResources        `json:"resources"`
	Labels          []string               `json:"labels"`
	Status          RunnerStatus           `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at"`
}

// GetNodeID returns the Node ID of the Runner or an empty string.
func (r Runner) GetNodeID() string {
	if r.NodeID == nil {
		return ""
	}

	return *r.NodeID
}

// Equals returns true if the given Runner is equal to the current Runner.
func (r Runner) Equals(other Runner) bool {
	return r.ID == other.ID && r.Status.State == other.Status.State
}

type RunnerAffinityRule struct {
	Key      string   `json:"key" mapstructure:"key"`
	Operator string   `json:"operator" mapstructure:"operator"`
	Values   []string `json:"values" mapstructure:"values"`
}

func (r *RunnerAffinityRule) String() string {
	return fmt.Sprintf("%s %s %v", r.Key, r.Operator, r.Values)
}

func (r *RunnerAffinityRule) Validate() error {
	var errs error

	if r.Key == "" {
		errs = multierror.Append(errs, fmt.Errorf("key is required"))
	}

	if r.Operator == "" {
		errs = multierror.Append(errs, fmt.Errorf("operator is required"))
	}

	if len(r.Values) == 0 {
		errs = multierror.Append(errs, fmt.Errorf("values is required"))
	}

	return errs
}

type RunnerImagePullPolicy string

func (r RunnerImagePullPolicy) String() string {
	return string(r)
}

const (
	RunnerImagePullPolicyAlways       RunnerImagePullPolicy = "Always"
	RunnerImagePullPolicyIfNotPresent RunnerImagePullPolicy = "IfNotPresent"
	RunnerImagePullPolicyNever        RunnerImagePullPolicy = "Never"
)

type RunnerState string

func (r RunnerState) String() string {
	return string(r)
}

const (
	RunnerStatePending   RunnerState = "Pending"
	RunnerStateIdle      RunnerState = "Idle"
	RunnerStateActive    RunnerState = "Active"
	RunnerStateCompleted RunnerState = "Completed"
)

type RunnerStatus struct {
	State       RunnerState `json:"state"`
	Description string      `json:"description"`
}

func (r *RunnerStatus) String() string {
	return r.State.String()
}

// RunnerFilterFunc is a function that filters a Runner. It returns true if the
// Runner should be included in the result set.
type RunnerFilterFunc func(*Runner) bool

// RunnerResources represents the resources required by a Runner. It is used to
// calculate the amount of resources that a Node needs to have available in
// order to run the Runner.
type RunnerResources struct {
	VCPUs    int64 `json:"vcpus" mapstructure:"vcpus"`
	MemoryMB int64 `json:"memory_mb" mapstructure:"memory_mb"`
}

// Validate validates the configuration.
func (c *RunnerResources) Validate() error {
	var errs error

	if c.VCPUs <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("vcpus must be greater than 0"))
	}

	if c.MemoryMB <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("memory_mb must be greater than 0"))
	}

	return errs
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
	ReconcileInterval time.Duration     `json:"reconcile_interval"`
	LastReconcileAt   time.Time         `json:"last_reconcile_at"`
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
func (r *NodeResource) Available() float64 {
	return float64(r.Capacity)*r.OvercommitRatio - float64(r.Allocated)
}

// String returns a string representation of the resource.
func (r *NodeResource) String() string {
	return fmt.Sprintf("%d/%d", r.Allocated, r.Capacity)
}

// IsAvailable returns true if the given amount of resource is available. It
// takes the overcommit ratio into account.
func (r *NodeResource) IsAvailable(amount int64) bool {
	return float64(amount) <= r.Available()
}

// IsFull returns true if the resource is fully allocated.
func (r *NodeResource) IsFull() bool {
	return r.Allocated >= r.Capacity
}
