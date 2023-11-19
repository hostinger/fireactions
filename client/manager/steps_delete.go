package manager

import (
	"context"
	"fmt"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/agent"
	"github.com/hostinger/fireactions/client/containerd"
	"github.com/hostinger/fireactions/client/microvm"
	"github.com/hostinger/fireactions/pkg/planner"
)

type deleteContainerImageSnapshotStep struct {
	containerd containerd.Client
	snapshotID string
}

func newDeleteContainerImageSnapshotStep(containerd containerd.Client, snapshotID string) *deleteContainerImageSnapshotStep {
	step := &deleteContainerImageSnapshotStep{
		containerd: containerd,
		snapshotID: snapshotID,
	}

	return step
}

func (s *deleteContainerImageSnapshotStep) Name() string {
	return "delete_container_image_snapshot"
}

func (s *deleteContainerImageSnapshotStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	var err error
	err = containerd.DeleteLease(ctx, s.containerd, s.snapshotID)
	if err != nil {
		return nil, err
	}

	err = containerd.RemoveSnapshot(ctx, s.containerd, "devmapper", s.snapshotID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type stopMicroVMStep struct {
	driver microvm.Driver
	runner *fireactions.Runner
}

func newStopMicroVMStep(driver microvm.Driver, runner *fireactions.Runner) *stopMicroVMStep {
	step := &stopMicroVMStep{
		driver: driver,
		runner: runner,
	}

	return step
}

func (s *stopMicroVMStep) Name() string {
	return "stop_microvm"
}

func (s *stopMicroVMStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	err := s.driver.StopVM(ctx, s.runner.ID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type deleteMicroVMStep struct {
	driver microvm.Driver
	runner *fireactions.Runner
}

func newDeleteMicroVMStep(driver microvm.Driver, runner *fireactions.Runner) *deleteMicroVMStep {
	step := &deleteMicroVMStep{
		driver: driver,
		runner: runner,
	}

	return step
}

func (s *deleteMicroVMStep) Name() string {
	return "delete_microvm"
}

func (s *deleteMicroVMStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	err := s.driver.DeleteVM(ctx, s.runner.ID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type stopAgentStep struct {
	driver microvm.Driver
	client fireactions.Client
	runner *fireactions.Runner
}

func newStopAgentStep(client fireactions.Client, driver microvm.Driver, runner *fireactions.Runner) *stopAgentStep {
	step := &stopAgentStep{
		driver: driver,
		runner: runner,
		client: client,
	}

	return step
}

func (s *stopAgentStep) Name() string {
	return "stop_agent"
}

func (s *stopAgentStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	microvm, err := s.driver.GetVM(ctx, s.runner.ID)
	if err != nil {
		return nil, err
	}

	runnerRemoveToken, _, err := s.client.GetRunnerRemoveToken(ctx, s.runner.ID)
	if err != nil {
		return nil, err
	}

	_, err = agent.NewClient(fmt.Sprintf("http://%s:6969", microvm.Status.IP)).Stop(ctx, &agent.StopRequest{
		Token: runnerRemoveToken.Token,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type deleteRunnerStep struct {
	client   fireactions.Client
	runnerID string
}

func newDeleteRunnerStep(client fireactions.Client, runnerID string) *deleteRunnerStep {
	step := &deleteRunnerStep{
		client:   client,
		runnerID: runnerID,
	}

	return step
}

func (s *deleteRunnerStep) Name() string {
	return "delete_runner"
}

func (s *deleteRunnerStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	_, err := s.client.DeleteRunner(ctx, s.runnerID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
