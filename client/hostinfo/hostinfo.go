package hostinfo

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/host"
)

// HostInfo is a struct that contains information about the host.
type HostInfo struct {
	Hostname  string
	MachineID string
	OS        string
	CpuInfo   *CpuInfo
	MemInfo   *MemInfo
	Uptime    uint64
}

// Collector is an interface that describes a host info collector.
type Collector interface {
	// Collect collects host info.
	Collect(ctx context.Context) (*HostInfo, error)

	// Last returns the last collected host info.
	Last() *HostInfo
}

type hostInfoCollector struct {
	lastHostInfo *HostInfo

	cpuInfoCollector *cpuInfoCollector
	memInfoCollector *memInfoCollector

	logger *zerolog.Logger
	l      *sync.Mutex
}

// NewCollector creates a new Collector.
func NewCollector(logger zerolog.Logger) *hostInfoCollector {
	logger = logger.With().Str("component", "host-info-collector").Logger()
	c := &hostInfoCollector{
		lastHostInfo:     &HostInfo{},
		cpuInfoCollector: newCpuInfoCollector(),
		memInfoCollector: newMemInfoCollector(),
		logger:           &logger,
		l:                &sync.Mutex{},
	}

	return c
}

// Collect collects host info.
func (c *hostInfoCollector) Collect(ctx context.Context) (*HostInfo, error) {
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

	err = c.collectMachineID()
	if err != nil {
		return nil, err
	}

	return c.lastHostInfo, nil
}

// Last returns the last collected host info.
func (c *hostInfoCollector) Last() *HostInfo {
	c.l.Lock()
	defer c.l.Unlock()

	return c.lastHostInfo
}

func (c *hostInfoCollector) collectMachineID() error {
	f, err := os.Open("/var/lib/dbus/machine-id")
	if err != nil {
		return err
	}
	defer f.Close()

	uuid, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	c.lastHostInfo.MachineID = strings.TrimSpace(string(uuid))
	return nil
}
