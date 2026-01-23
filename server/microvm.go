package server

import (
	"context"
	"time"
)

type MicroVM struct {
	VMID      string
	Pool      string
	IPAddr    string
	CreatedAt time.Time
}

// ListMicroVMs retrieves the names and IP addresses of all running Firecracker VMs inside the pool.
func (p *Pool) ListMicroVMs(ctx context.Context, pool string) ([]*MicroVM, error) {
	p.machinesMu.Lock()
	defer p.machinesMu.Unlock()

	microVMs := make([]*MicroVM, 0, len(p.machines))

	for _, metadata := range p.machines {
		ipAddr := ""
		if len(metadata.machine.Cfg.NetworkInterfaces) > 0 {
			ipAddr = metadata.machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String()
		}

		vm := &MicroVM{
			VMID:      metadata.machine.Cfg.VMID,
			Pool:      p.config.Name,
			IPAddr:    ipAddr,
			CreatedAt: metadata.createdAt,
		}

		microVMs = append(microVMs, vm)
	}

	return microVMs, nil
}
