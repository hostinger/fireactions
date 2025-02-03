package server

import "context"

type MicroVM struct {
	VMID string
}

// ListMicroVMs retrieves the names of all running Firecracker VMs inside the pool.
func (p *Pool) ListMicroVMs(ctx context.Context, pool string) ([]*MicroVM, error) {
	p.machinesMu.Lock()
	defer p.machinesMu.Unlock()

	microVMs := make([]*MicroVM, 0, len(p.machines))

	for _, machine := range p.machines {
		vm := &MicroVM{
			VMID: machine.Cfg.VMID,
		}
		microVMs = append(microVMs, vm)
	}

	return microVMs, nil
}
