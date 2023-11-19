package manager

import (
	"context"
	"fmt"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/containerd"
	"github.com/hostinger/fireactions/client/microvm"
	"github.com/hostinger/fireactions/pkg/planner"
)

type createGitHubRunnerPlan struct {
	driver     microvm.Driver
	client     fireactions.Client
	containerd containerd.Client
	runner     *fireactions.Runner
}

func newCreateGitHubRunnerPlan(
	containerd containerd.Client, client fireactions.Client, driver microvm.Driver, runner *fireactions.Runner) *createGitHubRunnerPlan {
	plan := &createGitHubRunnerPlan{
		driver:     driver,
		runner:     runner,
		containerd: containerd,
		client:     client,
	}

	return plan
}

func (p *createGitHubRunnerPlan) Name() string {
	return "create_github_runner"
}

func (p *createGitHubRunnerPlan) Create(ctx context.Context) ([]planner.Procedure, error) {
	steps := []planner.Procedure{}

	switch p.runner.ImagePullPolicy {
	case fireactions.RunnerImagePullPolicyIfNotPresent:
		ok, err := containerd.ImageExists(ctx, p.containerd, p.runner.Image)
		if err != nil {
			return nil, err
		}

		if ok {
			break
		}

		steps = append(steps, newPullContainerImageStep(p.containerd, p.runner.Image))
	case fireactions.RunnerImagePullPolicyNever:
		ok, err := containerd.ImageExists(ctx, p.containerd, p.runner.Image)
		if err != nil {
			return nil, err
		}

		if ok {
			break
		}

		return nil, fmt.Errorf("image %s does not exist and image pull policy is never", p.runner.Image)
	case fireactions.RunnerImagePullPolicyAlways:
		steps = append(steps, newPullContainerImageStep(p.containerd, p.runner.Image))
	default:
		return nil, fmt.Errorf("unknown image pull policy %s", p.runner.ImagePullPolicy)
	}

	snapshotID := fmt.Sprintf("fireactions/runner/%s", p.runner.ID)
	ok, err := containerd.SnapshotExists(ctx, p.containerd, "devmapper", snapshotID)
	if err != nil {
		return nil, fmt.Errorf("containerd: getting snapshot: %w", err)
	}

	if !ok {
		steps = append(steps, newCreateContainerImageSnapshotStep(p.containerd, p.runner.Image, snapshotID))
	}

	mvm, err := p.driver.GetVM(ctx, p.runner.ID)
	if err != nil {
		if err != microvm.ErrNotFound {
			return nil, err
		}

		steps = append(steps,
			newCreateMicroVMStep(p.driver, p.containerd, p.runner),
			newStartMicroVMStep(p.driver, p.runner),
			newStartAgentStep(p.client, p.driver, p.runner))
	} else {
		if mvm.Status.State != microvm.MicroVMStateRunning {
			steps = append(steps, newStartMicroVMStep(p.driver, p.runner), newStartAgentStep(p.client, p.driver, p.runner))
		}
	}

	if p.runner.Status.Phase == fireactions.RunnerPhasePending {
		steps = append(steps, newSetRunnerStatusStep(p.client, p.runner.ID, fireactions.RunnerPhaseIdle))
	}

	return steps, nil
}

type deleteGitHubRunnerPlan struct {
	driver     microvm.Driver
	client     fireactions.Client
	containerd containerd.Client
	runner     *fireactions.Runner
}

func newDeleteGitHubRunnerPlan(
	containerd containerd.Client, client fireactions.Client, driver microvm.Driver, runner *fireactions.Runner) *deleteGitHubRunnerPlan {
	plan := &deleteGitHubRunnerPlan{
		driver:     driver,
		runner:     runner,
		containerd: containerd,
		client:     client,
	}

	return plan
}

func (p *deleteGitHubRunnerPlan) Name() string {
	return "delete_github_runner"
}

func (p *deleteGitHubRunnerPlan) Create(ctx context.Context) ([]planner.Procedure, error) {
	steps := []planner.Procedure{}

	mvm, err := p.driver.GetVM(ctx, p.runner.ID)
	if err == nil {
		if mvm.Status.State == microvm.MicroVMStateRunning {
			steps = append(steps, newStopAgentStep(p.client, p.driver, p.runner), newStopMicroVMStep(p.driver, p.runner))
		}

		steps = append(steps, newDeleteMicroVMStep(p.driver, p.runner))
	}

	snapshotID := fmt.Sprintf("fireactions/runner/%s", p.runner.ID)
	ok, err := containerd.SnapshotExists(ctx, p.containerd, "devmapper", snapshotID)
	if err != nil {
		return nil, fmt.Errorf("containerd: getting snapshot: %w", err)
	}

	if ok {
		steps = append(steps, newDeleteContainerImageSnapshotStep(p.containerd, snapshotID))
	}

	steps = append(steps, newDeleteRunnerStep(p.client, p.runner.ID))
	return steps, nil
}
