package server

import (
	"github.com/hostinger/fireactions"
)

func convertPool(p *Pool) *fireactions.Pool {
	pool := &fireactions.Pool{
		Name:       p.config.Name,
		MaxRunners: p.config.MaxRunners,
		MinRunners: p.config.MinRunners,
		CurRunners: p.GetCurrentSize(),
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
		VMID:   m.VMID,
		IPAddr: m.IPAddr,
	}

	return microVM
}
