package main

import (
	"strings"
	"time"

	"github.com/docker/go-units"
	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
)

// printablePool wraps a proto Pool for printing
type printablePool struct {
	Pools []*serverv1.Pool
}

func (p *printablePool) Cols() []string {
	return []string{"Name", "Current", "Desired", "Organization", "Group ID", "Labels", "Image", "State"}
}

func (p *printablePool) ColsMap() map[string]string {
	return map[string]string{
		"Name":         "Name",
		"Current":      "Current",
		"Desired":      "Desired",
		"Organization": "Organization",
		"Group ID":     "Group ID",
		"Labels":       "Labels",
		"Image":        "Image",
		"State":        "State",
	}
}

func (p *printablePool) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(p.Pools))
	for _, pool := range p.Pools {
		state := "Active"
		if pool.State == serverv1.PoolState_POOL_STATE_PAUSED {
			state = "Paused"
		}
		kv = append(kv, map[string]interface{}{
			"Name":         pool.Name,
			"Current":      pool.CurrentReplicas,
			"Desired":      pool.DesiredReplicas,
			"Organization": pool.Organization,
			"Group ID":     pool.GroupId,
			"Labels":       strings.Join(pool.Labels, ", "),
			"Image":        pool.Image,
			"State":        state,
		})
	}
	return kv
}

// printableMachine wraps a slice of proto Machines for printing
type printableMachine struct {
	Machines []*serverv1.Machine
}

func (m *printableMachine) Cols() []string {
	return []string{"Pool", "ID", "ADDR", "Runner State", "Runner Version", "Created"}
}

func (m *printableMachine) ColsMap() map[string]string {
	return map[string]string{
		"Pool":           "Pool",
		"ID":             "ID",
		"ADDR":           "ADDR",
		"Runner State":   "Runner State",
		"Runner Version": "Runner Version",
		"Created":        "Created",
	}
}

func (m *printableMachine) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(m.Machines))
	for _, vm := range m.Machines {
		runnerState := vm.RunnerState
		if runnerState == "" {
			runnerState = "Unknown"
		}

		runnerVersion := vm.RunnerVersion
		if runnerVersion == "" {
			runnerVersion = "Unknown"
		}

		createdAt := vm.CreatedAt.AsTime()

		kv = append(kv, map[string]interface{}{
			"Pool":           vm.Pool,
			"ID":             vm.ID,
			"ADDR":           vm.Addr,
			"Runner State":   runnerState,
			"Runner Version": runnerVersion,
			"Created":        units.HumanDuration(time.Since(createdAt)),
		})
	}
	return kv
}

// printableImage wraps a slice of proto Images for printing
type printableImage struct {
	Images []*serverv1.Image
}

func (i *printableImage) Cols() []string {
	return []string{"Name", "Size", "Created"}
}

func (i *printableImage) ColsMap() map[string]string {
	return map[string]string{
		"Name":    "Name",
		"Size":    "Size",
		"Created": "Created",
	}
}

func (i *printableImage) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(i.Images))
	for _, img := range i.Images {
		createdAt := img.CreatedAt.AsTime()

		kv = append(kv, map[string]interface{}{
			"Name":    img.Name,
			"Size":    units.HumanSize(float64(img.GetSize())),
			"Created": units.HumanDuration(time.Since(createdAt)),
		})
	}
	return kv
}
