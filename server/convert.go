package server

import (
	"github.com/hostinger/fireactions"
)

func convertPool(p *Pool) *fireactions.Pool {
	replicas := p.GetReplicas()
	pool := &fireactions.Pool{
		Name:            p.config.Name,
		Replicas:        replicas,
		CurrentReplicas: p.GetCurrentSize(),
		DesiredReplicas: replicas,
		Organization:    p.config.Runner.Organization,
		GroupID:         p.config.Runner.GroupID,
		Labels:          p.config.Runner.Labels,
		Image:           p.config.Runner.Image,
	}

	if p.isActive {
		pool.Status = fireactions.PoolStatus{
			State:   fireactions.PoolStateActive,
			Message: "Pool is active",
		}
	} else {
		pool.Status = fireactions.PoolStatus{
			State:   fireactions.PoolStatePaused,
			Message: "Pool is paused",
		}
	}

	return pool
}

func convertPools(pools []*Pool) fireactions.Pools {
	convertedPools := make(fireactions.Pools, 0, len(pools))
	for _, pool := range pools {
		convertedPools = append(convertedPools, convertPool(pool))
	}

	return convertedPools
}

func convertMicroVM(m *MicroVM) *fireactions.MicroVM {
	microVM := &fireactions.MicroVM{
		VMID:      m.VMID,
		Pool:      m.Pool,
		IPAddr:    m.IPAddr,
		CreatedAt: m.CreatedAt,
	}

	return microVM
}

func convertMicroVMs(microVMs []*MicroVM) fireactions.MicroVMs {
	convertedMicroVMs := make(fireactions.MicroVMs, 0, len(microVMs))
	for _, microVM := range microVMs {
		convertedMicroVMs = append(convertedMicroVMs, *convertMicroVM(microVM))
	}

	return convertedMicroVMs
}
