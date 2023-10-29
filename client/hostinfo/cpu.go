package hostinfo

import (
	"context"

	"github.com/shirou/gopsutil/cpu"
)

// CpuInfo is a struct that contains information about the CPU.
type CpuInfo struct {
	NumCores int
}

type cpuInfoCollector struct {
	cpuInfo *CpuInfo
}

// Collect collects information about the CPU.
func (c *cpuInfoCollector) Collect(ctx context.Context) (*CpuInfo, error) {
	cpu, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var numCores int
	for _, c := range cpu {
		numCores += int(c.Cores)
	}
	c.cpuInfo.NumCores = numCores

	return c.cpuInfo, nil
}

func newCpuInfoCollector() *cpuInfoCollector {
	c := &cpuInfoCollector{
		cpuInfo: &CpuInfo{},
	}

	return c
}
