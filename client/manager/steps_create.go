package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/agent"
	"github.com/hostinger/fireactions/client/containerd"
	"github.com/hostinger/fireactions/client/microvm"
	"github.com/hostinger/fireactions/pkg/planner"
)

type createContainerImageSnapshotStep struct {
	containerd containerd.Client
	imageRef   string
	snapshotID string
}

func newCreateContainerImageSnapshotStep(containerd containerd.Client, imageRef, snapshotID string) *createContainerImageSnapshotStep {
	step := &createContainerImageSnapshotStep{
		containerd: containerd,
		imageRef:   imageRef,
		snapshotID: snapshotID,
	}

	return step
}

func (s *createContainerImageSnapshotStep) Name() string {
	return "create_container_image_snapshot"
}

func (s *createContainerImageSnapshotStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	leaseCtx, err := containerd.NewContextWithLease(ctx, s.containerd, s.snapshotID)
	if err != nil {
		return nil, err
	}

	err = containerd.CreateSnapshot(leaseCtx, s.containerd, s.imageRef, "devmapper", s.snapshotID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type pullContainerImageStep struct {
	containerd containerd.Client
	imageRef   string
}

func newPullContainerImageStep(containerd containerd.Client, imageRef string) *pullContainerImageStep {
	step := &pullContainerImageStep{
		containerd: containerd,
		imageRef:   imageRef,
	}

	return step
}

func (s *pullContainerImageStep) Name() string {
	return "pull_container_image"
}

func (s *pullContainerImageStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	return nil, containerd.PullImage(ctx, s.containerd, s.imageRef)
}

type createMicroVMStep struct {
	driver     microvm.Driver
	runner     *fireactions.Runner
	containerd containerd.Client
}

func newCreateMicroVMStep(driver microvm.Driver, containerd containerd.Client, runner *fireactions.Runner) *createMicroVMStep {
	step := &createMicroVMStep{
		driver:     driver,
		runner:     runner,
		containerd: containerd,
	}

	return step
}

func (s *createMicroVMStep) Name() string {
	return "create_microvm"
}

func (s *createMicroVMStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	snapshotID := fmt.Sprintf("fireactions/runner/%s", s.runner.ID)
	mounts, err := containerd.GetSnapshotMounts(ctx, s.containerd, "devmapper", snapshotID)
	if err != nil {
		return nil, fmt.Errorf("containerd: getting snapshot mounts: %w", err)
	}

	rootDrivePath := mounts[0].Source
	s.runner.Metadata["fireactions"] = map[string]interface{}{
		"runner_id": s.runner.ID,
	}

	err = s.driver.CreateVM(ctx, &microvm.MicroVM{ID: s.runner.ID, Status: microvm.MicroVMStatus{State: microvm.MicroVMStateUnknown}, Spec: microvm.MicroVMSpec{
		Name:              s.runner.Name,
		VCPU:              s.runner.Resources.VCPUs,
		MemoryBytes:       s.runner.Resources.MemoryBytes,
		Metadata:          s.runner.Metadata,
		Drives:            []microvm.Drive{{ID: "rootfs", PathOnHost: rootDrivePath, IsRoot: true, IsReadOnly: false}},
		NetworkInterfaces: []microvm.NetworkInterface{{AllowMMDS: true, NetworkName: "fireactions", IfName: "veth0"}},
	}})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type startMicroVMStep struct {
	driver microvm.Driver
	runner *fireactions.Runner
}

func newStartMicroVMStep(driver microvm.Driver, runner *fireactions.Runner) *startMicroVMStep {
	step := &startMicroVMStep{
		driver: driver,
		runner: runner,
	}

	return step
}

func (s *startMicroVMStep) Name() string {
	return "start_microvm"
}

func (s *startMicroVMStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	err := s.driver.StartVM(ctx, s.runner.ID)
	if err != nil {
		return nil, err
	}

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
		}

		microvm, err := s.driver.GetVM(ctx, s.runner.ID)
		if err != nil {
			return nil, err
		}

		_, err = agent.NewClient(fmt.Sprintf("http://%s:6969", microvm.Status.IP)).Ping(ctx)
		if err != nil {
			continue
		}

		break
	}

	return nil, nil
}

type startAgentStep struct {
	driver microvm.Driver
	client fireactions.Client
	runner *fireactions.Runner
}

func newStartAgentStep(client fireactions.Client, driver microvm.Driver, runner *fireactions.Runner) *startAgentStep {
	step := &startAgentStep{
		driver: driver,
		runner: runner,
		client: client,
	}

	return step
}

func (s *startAgentStep) Name() string {
	return "start_agent"
}

func (s *startAgentStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	microvm, err := s.driver.GetVM(ctx, s.runner.ID)
	if err != nil {
		return nil, err
	}

	runnerRegistrationToken, _, err := s.client.GetRunnerRegistrationToken(ctx, s.runner.ID)
	if err != nil {
		return nil, err
	}

	_, err = agent.NewClient(fmt.Sprintf("http://%s:6969", microvm.Status.IP)).Start(ctx, &agent.StartRequest{
		Name:          s.runner.Name,
		URL:           fmt.Sprintf("https://github.com/%s", s.runner.Organisation),
		Token:         runnerRegistrationToken.Token,
		Labels:        s.runner.Labels,
		Ephemeral:     true,
		DisableUpdate: true,
		Replace:       true,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type setRunnerStatusStep struct {
	client      fireactions.Client
	runnerPhase fireactions.RunnerPhase
	runnerID    string
}

func newSetRunnerStatusStep(client fireactions.Client, runnerID string, runnerPhase fireactions.RunnerPhase) *setRunnerStatusStep {
	step := &setRunnerStatusStep{
		client:      client,
		runnerPhase: runnerPhase,
		runnerID:    runnerID,
	}

	return step
}

func (s *setRunnerStatusStep) Name() string {
	return "set_runner_status"
}

func (s *setRunnerStatusStep) Do(ctx context.Context) ([]planner.Procedure, error) {
	_, err := s.client.SetRunnerStatus(ctx, s.runnerID, fireactions.SetRunnerStatusRequest{
		Phase: s.runnerPhase,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
