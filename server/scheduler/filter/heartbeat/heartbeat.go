package heartbeat

import (
	"context"
	"fmt"
	"time"

	timeago "github.com/caarlos0/timea.go"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler/filter"
)

// Filter implements filter.Filter interface.
type Filter struct {
}

// New creates new Filter.
func New() *Filter {
	return &Filter{}
}

var _ filter.Filter = &Filter{}

// Name returns the name of the filter.
func (f *Filter) Name() string {
	return "heartbeat"
}

// Filter filters out nodes that didn't send a heartbeat for more than the configured heartbeat interval.
func (f *Filter) Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	if time.Since(node.LastHeartbeat) > node.HeartbeatInterval {
		return false, fmt.Errorf("node is not alive: last heartbeat was %s", timeago.Of(node.LastHeartbeat))
	}

	return true, nil
}
