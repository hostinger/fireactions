package server

import (
	"context"
	"time"

	"github.com/containerd/containerd"
	agentv1 "github.com/hostinger/fireactions/proto/agent/v1"
	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// convertPoolToProto converts a Pool to its protobuf representation.
func convertPoolToProto(ctx context.Context, pool *Pool) *serverv1.Pool {
	state := serverv1.PoolState_POOL_STATE_ACTIVE
	if !pool.isActive {
		state = serverv1.PoolState_POOL_STATE_PAUSED
	}

	return &serverv1.Pool{
		Name:            pool.config.Name,
		Organization:    pool.config.Runner.Organization,
		Replicas:        int32(pool.GetReplicas()),
		CurrentReplicas: int32(pool.GetCurrentSize()),
		DesiredReplicas: int32(pool.GetReplicas()),
		GroupId:         pool.config.Runner.GroupID,
		Labels:          pool.config.Runner.Labels,
		Image:           pool.config.Runner.Image,
		State:           state,
	}
}

func convertMachineToProto(ctx context.Context, machine *Machine) *serverv1.Machine {
	m := &serverv1.Machine{
		ID:        machine.Name,
		Pool:      machine.Pool,
		Addr:      machine.GetAddr(),
		CreatedAt: timestamppb.New(machine.CreatedAt),
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, client, err := machine.ConnectToGuestAgent(ctx)
	if err != nil {
		m.RunnerState = "Unknown"
		m.RunnerVersion = "Unknown"
		return m
	}
	defer conn.Close()

	runnerStateResp, err := client.GetRunnerState(
		ctx, &agentv1.GetRunnerStateRequest{})
	if err != nil {
		m.RunnerState = "Unknown"
	} else {
		m.RunnerState = runnerStateResp.GetState()
	}

	runnerVersionResp, err := client.GetRunnerVersion(
		ctx, &agentv1.GetRunnerVersionRequest{})
	if err != nil {
		m.RunnerVersion = "Unknown"
	} else {
		m.RunnerVersion = runnerVersionResp.GetVersion()
	}

	return m
}

// convertImageToProto converts a containerd Image to its protobuf representation.
func convertImageToProto(ctx context.Context, img containerd.Image) *serverv1.Image {
	size, _ := img.Size(ctx)
	createdAt := img.Metadata().CreatedAt

	i := &serverv1.Image{
		Name:      img.Name(),
		Size:      size,
		CreatedAt: timestamppb.New(createdAt),
	}

	return i
}
