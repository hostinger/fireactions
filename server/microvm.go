package server

import (
	"context"
	"fmt"
	"strings"
)

type MicroVM struct {
	VMID   string
	IPAddr string
}

// ListMicroVMs retrieves the names and IP addresses of all running Firecracker VMs inside the pool.
func (p *Pool) ListMicroVMs(ctx context.Context, pool string) ([]*MicroVM, error) {
	p.machinesMu.Lock()
	defer p.machinesMu.Unlock()

	microVMs := make([]*MicroVM, 0, len(p.machines))

	for _, machine := range p.machines {
		vm := &MicroVM{
			VMID:   machine.Cfg.VMID,
			IPAddr: machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP.String(),
		}

		microVMs = append(microVMs, vm)
	}
	return microVMs, nil
}

// ExtractPoolNameFromVMID parses a VMID to get the associated Pool name.
// VMIDs (runnerName) are formatted as "pool-suffix", so this avoids scanning all Pools.
func ExtractPoolNameFromVMID(vmid string) (string, error) {
	// Locate the last '-' and verify it is neither the first nor last character, ensuring both the pool name and suffix are non-empty.
	if i := strings.LastIndex(vmid, "-"); i > 0 && i < len(vmid)-1 {
		return vmid[:i], nil
	}
	return "", fmt.Errorf("invalid VMID format: %q", vmid)
}
