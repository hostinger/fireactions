package server

import (
	"context"
	"fmt"
	"time"

	githubv50 "github.com/google/go-github/v50/github"
	githubv53 "github.com/google/go-github/v53/github"
	"github.com/hostinger/fireactions"
)

func (s *Server) createOrUpdateWorkflowRun(deliveryID string, eventName string, event *githubv50.WorkflowRunEvent) error {
	logger := s.logger.With().Str("delivery_id", deliveryID).Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.store.SaveWorkflowRun(ctx, event.GetOrg().GetLogin(), event.GetWorkflowRun())
	if err != nil {
		logger.Error().Err(err).Msgf("failed to save workflow run %d", event.GetWorkflowRun().GetID())
		return err
	}

	logger.Debug().Msgf("saved workflow run %d", event.GetWorkflowRun().GetID())
	return nil
}

func (s *Server) createOrUpdateWorkflowJob(deliveryID string, eventName string, event *githubv50.WorkflowJobEvent) error {
	logger := s.logger.With().Str("delivery_id", deliveryID).Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.store.SaveWorkflowJob(ctx, event.GetOrg().GetLogin(), event.GetWorkflowJob())
	if err != nil {
		logger.Error().Err(err).Msgf("failed to save workflow job %d", event.GetWorkflowJob().GetID())
		return err
	}

	logger.Debug().Msgf("saved workflow job %d", event.GetWorkflowJob().GetID())
	return nil
}

func (s *Server) updateRunnerForWorkflowJob(deliveryID string, eventName string, event *githubv50.WorkflowJobEvent) error {
	logger := s.logger.With().Str("delivery_id", deliveryID).Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, ok := s.config.FindMatchingJobLabel(event.GetWorkflowJob().Labels)
	if !ok {
		logger.Debug().Msgf("skipped updating GitHub runner for workflow job %d: no matching label found", event.GetWorkflowJob().GetID())
		return nil
	}

	runners, err := s.store.GetRunners(ctx, func(runner *fireactions.Runner) bool {
		return runner.Name == event.GetWorkflowJob().GetRunnerName()
	})
	if err != nil {
		return err
	}

	if len(runners) == 0 {
		return nil
	}

	var status fireactions.RunnerStatus
	switch event.GetWorkflowJob().GetStatus() {
	case "in_progress":
		status = fireactions.RunnerStatus{State: fireactions.RunnerStateActive, Description: fmt.Sprintf("Job %d is in progress", event.GetWorkflowJob().GetID())}
	case "completed":
		status = fireactions.RunnerStatus{State: fireactions.RunnerStateCompleted, Description: fmt.Sprintf("Job %d is completed", event.GetWorkflowJob().GetID())}
	default:
		return nil
	}

	runner := runners[0]
	_, err = s.store.UpdateRunner(ctx, runner.ID, func(r *fireactions.Runner) error {
		r.Status = status
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info().Msgf("updated GitHub runner %s for workflow job %d", event.GetWorkflowJob().GetRunnerName(), event.GetWorkflowJob().GetID())
	return nil
}

func (s *Server) createRunnerForWorkflowJob(deliveryID string, eventName string, event *githubv50.WorkflowJobEvent) error {
	logger := s.logger.With().Str("delivery_id", deliveryID).Logger()

	jobLabel, ok := s.config.FindMatchingJobLabel(event.GetWorkflowJob().Labels)
	if !ok {
		logger.Debug().Msgf("skipped creating GitHub runner for workflow job %d: no matching label found", event.GetWorkflowJob().GetID())
		return nil
	}

	if !jobLabel.IsAllowedRepository(event.GetRepo().GetFullName()) {
		logger.Debug().Msgf("skipped creating GitHub runner for workflow job %d: repository is not allowed", event.GetWorkflowJob().GetID())
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if event.GetRepo().GetOwner().GetType() != "Organization" {
		logger.Debug().Msgf("skipped creating GitHub runner for workflow job %d: repository owner is not an organization", event.GetWorkflowJob().GetID())
		return nil
	}

	runner := convertJobLabelConfigToRunner(jobLabel, event.GetRepo().GetOwner().GetLogin())

	client := s.github.InstallationClient(event.GetInstallation().GetID())
	config, _, err := client.Actions.GenerateOrgJITConfig(ctx, runner.Organisation, &githubv53.GenerateJITConfigRequest{
		Name:          runner.Name,
		RunnerGroupID: 1,
		Labels:        runner.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to generate JIT config: %s", err.Error())
	}
	runner.Metadata["fireactions"] = map[string]interface{}{
		"runner_id":         runner.ID,
		"runner_jit_config": config.GetEncodedJITConfig(),
	}

	if err := s.store.SaveRunner(ctx, runner); err != nil {
		return fmt.Errorf("failed to save runner: %s", err.Error())
	}

	s.scheduler.AddToQueue(runner)
	logger.Info().Msgf("created GitHub runner %s for workflow job %d", runner.Name, event.WorkflowJob.GetID())
	return nil
}
