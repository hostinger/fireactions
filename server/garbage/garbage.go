package garbage

import "context"

// Collector is an interface for garbage collectors.
type Collector interface {
	Run(ctx context.Context)
	GarbageCollect(ctx context.Context) error
}
