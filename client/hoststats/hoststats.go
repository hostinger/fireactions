package hoststats

import (
	"context"
	"sync"

	"github.com/shirou/gopsutil/host"
)

// HostInfo is a struct that contains information about the host.
type HostInfo struct {
	Hostname string
	OS       string
	CpuInfo  *CpuInfo
	MemInfo  *MemInfo
	Uptime   uint64
}

// Collector is an interface that describes a host info collector.
type Collector interface {
	// Collect collects host info.
	Collect(ctx context.Context) (*HostInfo, error)

	// Last returns the last collected host info.
	Last() *HostInfo
}

type hostStatsCollector struct {
	lastHostInfo *HostInfo

	cpuInfoCollector *cpuInfoCollector
	memInfoCollector *memInfoCollector

	l *sync.Mutex
}

// NewCollector creates a new Collector.
func NewCollector() *hostStatsCollector {
	c := &hostStatsCollector{
		lastHostInfo:     &HostInfo{},
		cpuInfoCollector: newCpuInfoCollector(),
		memInfoCollector: newMemInfoCollector(),
		l:                &sync.Mutex{},
	}

	return c
}

// Collect collects host info.
func (c *hostStatsCollector) Collect(ctx context.Context) (*HostInfo, error) {
	c.l.Lock()
	defer c.l.Unlock()

	var err error

	host, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	c.lastHostInfo.Hostname = host.Hostname
	c.lastHostInfo.OS = host.OS
	c.lastHostInfo.Uptime = host.Uptime

	c.lastHostInfo.CpuInfo, err = c.cpuInfoCollector.Collect(ctx)
	if err != nil {
		return nil, err
	}

	c.lastHostInfo.MemInfo, err = c.memInfoCollector.Collect(ctx)
	if err != nil {
		return nil, err
	}

	return c.lastHostInfo, nil
}

// Last returns the last collected host info.
func (c *hostStatsCollector) Last() *HostInfo {
	c.l.Lock()
	defer c.l.Unlock()

	return c.lastHostInfo
}
