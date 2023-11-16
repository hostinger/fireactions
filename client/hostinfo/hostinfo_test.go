package hostinfo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	collector := NewCollector()

	hostInfo, err := collector.Collect(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, hostInfo)
	assert.NotNil(t, hostInfo.CpuInfo)
	assert.NotNil(t, hostInfo.MemInfo)
	assert.NotNil(t, hostInfo.Hostname)
	assert.NotNil(t, hostInfo.OS)
	assert.NotNil(t, hostInfo.Uptime)

	lastHostInfo := collector.Last()

	assert.Equal(t, lastHostInfo.Uptime, hostInfo.Uptime)
}
