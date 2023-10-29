package hostinfo

import (
	"context"

	"github.com/shirou/gopsutil/mem"
)

// MemInfo is a struct that contains information about the RAM.
type MemInfo struct {
	Total uint64
}

type memInfoCollector struct {
	memInfo *MemInfo
}

// Collect collects information about the RAM.
func (c *memInfoCollector) Collect(ctx context.Context) (*MemInfo, error) {
	mem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	c.memInfo.Total = mem.Total

	return c.memInfo, nil
}

func newMemInfoCollector() *memInfoCollector {
	c := &memInfoCollector{
		memInfo: &MemInfo{},
	}

	return c
}
