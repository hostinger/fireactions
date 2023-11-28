package server

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/gin-gonic/gin"
	githubv50 "github.com/google/go-github/v50/github"
	githubv53 "github.com/google/go-github/v53/github"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/stringid"
	"github.com/samber/lo"
)

func (s *Server) handleWebhook() gin.HandlerFunc {
	handler := githubevents.New(s.config.GitHub.WebhookSecret)
	handler.OnWorkflowJobEventCompleted(s.handleWorkflowJobCompleted)
	handler.OnWorkflowJobEventQueued(s.handleWorkflowJobQueued)
	handler.OnWorkflowJobEventInProgress(s.handleWorkflowJobInProgress)

	f := func(ctx *gin.Context) {
		err := handler.HandleEventRequest(ctx.Request)
		if err != nil {
			s.logger.Error().Err(err).Msgf("failed to handle GitHub event")
			ctx.AbortWithStatus(500)
			return
		}

		ctx.Status(200)
	}

	return f
}

func (s *Server) handleWorkflowJobCompleted(deliveryID string, eventName string, event *githubv50.WorkflowJobEvent) error {
	logger := s.logger.With().Str("delivery_id", deliveryID).Logger()

	_, ok := s.findWorkflowJobLabel(event)
	if !ok {
		logger.Debug().Msgf("skipping workflow job %d: no matching label found", *event.WorkflowJob.ID)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	runners, err := s.store.GetRunners(ctx, func(runner *fireactions.Runner) bool {
		return runner.Name == event.GetWorkflowJob().GetRunnerName()
	})
	if err != nil {
		return err
	}

	if len(runners) == 0 {
		logger.Debug().Msgf("skipped updating GitHub runner %s for workflow job %d: not found", *event.WorkflowJob.RunnerName, *event.WorkflowJob.ID)
		return nil
	}

	runner := runners[0]
	_, err = s.store.UpdateRunner(ctx, runner.ID, func(r *fireactions.Runner) error {
		r.Status = fireactions.RunnerStatus{State: fireactions.RunnerStateCompleted, Description: fmt.Sprintf("Job %d is completed", *event.WorkflowJob.ID)}
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info().Msgf("updated GitHub runner %s for workflow job %d (job is completed)", *event.WorkflowJob.RunnerName, *event.WorkflowJob.ID)
	return nil
}

func (s *Server) handleWorkflowJobInProgress(deliveryID string, eventName string, event *githubv50.WorkflowJobEvent) error {
	logger := s.logger.With().Str("delivery_id", deliveryID).Logger()

	_, ok := s.findWorkflowJobLabel(event)
	if !ok {
		logger.Debug().Msgf("skipping workflow job %d: no matching label found", *event.WorkflowJob.ID)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	runners, err := s.store.GetRunners(ctx, func(runner *fireactions.Runner) bool {
		return runner.Name == event.GetWorkflowJob().GetRunnerName()
	})
	if err != nil {
		return err
	}

	if len(runners) == 0 {
		logger.Debug().Msgf("skipped updating GitHub runner %s for workflow job %d: not found", *event.WorkflowJob.RunnerName, *event.WorkflowJob.ID)
		return nil
	}

	runner := runners[0]
	_, err = s.store.UpdateRunner(ctx, runner.ID, func(r *fireactions.Runner) error {
		r.Status = fireactions.RunnerStatus{State: fireactions.RunnerStateActive, Description: fmt.Sprintf("Job %d is in progress", *event.WorkflowJob.ID)}
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info().Msgf("updated GitHub runner %s for workflow job %d (job is in progress)", *event.WorkflowJob.RunnerName, *event.WorkflowJob.ID)
	return nil
}

func (s *Server) handleWorkflowJobQueued(deliveryID string, eventName string, event *githubv50.WorkflowJobEvent) error {
	logger := s.logger.With().Str("delivery_id", deliveryID).Logger()

	jobLabel, ok := s.findWorkflowJobLabel(event)
	if !ok {
		logger.Debug().Msgf("skipping workflow job %d: no matching label found", *event.WorkflowJob.ID)
		return nil
	}

	validator := newWorkflowJobEventValidator(event, jobLabel)
	err := validator.Validate()
	if err != nil {
		logger.Debug().Msgf("skipping workflow job %d: %s", *event.WorkflowJob.ID, err.Error())
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if event.GetRepo().GetOwner().GetType() != "Organization" {
		logger.Debug().Msgf("skipping workflow job %d: repository owner is not an organization", *event.WorkflowJob.ID)
		return nil
	}

	runner := s.newRunnerFromJobLabelConfig(jobLabel, event.GetRepo().GetOwner().GetLogin())

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
	logger.Info().Msgf("created GitHub runner %s for workflow job %d", runner.Name, *event.WorkflowJob.ID)
	return nil
}

func (s *Server) findWorkflowJobLabel(event *githubv50.WorkflowJobEvent) (*JobLabelConfig, bool) {
	for _, label := range event.WorkflowJob.Labels {
		jobLabel, ok := s.config.FindJobLabel(label)
		if !ok {
			continue
		}

		return jobLabel, true
	}

	return nil, false
}

type workflowJobEventValidator struct {
	event  *githubv50.WorkflowJobEvent
	config *JobLabelConfig
}

func newWorkflowJobEventValidator(event *githubv50.WorkflowJobEvent, config *JobLabelConfig) *workflowJobEventValidator {
	w := &workflowJobEventValidator{event: event, config: config}
	return w
}

func (w *workflowJobEventValidator) Validate() error {
	if lo.ContainsBy(w.config.AllowedRepositories, func(repository string) bool {
		regexp, err := regexp.Compile(repository)
		if err != nil {
			return false
		}

		return regexp.MatchString(*w.event.Repo.FullName)
	}) {
		return nil
	}

	return fmt.Errorf("repository not allowed")
}

func (s *Server) newRunnerFromJobLabelConfig(config *JobLabelConfig, organisation string) *fireactions.Runner {
	runnerID := stringid.New()
	runner := &fireactions.Runner{
		ID:              runnerID,
		Organisation:    organisation,
		Name:            config.MustGetRunnerName(runnerID),
		NodeID:          nil,
		Labels:          config.GetRunnerLabels(),
		Resources:       config.RunnerResources,
		ImagePullPolicy: config.RunnerImagePullPolicy,
		Image:           config.RunnerImage,
		Metadata:        config.RunnerMetadata,
		Affinity:        config.RunnerAffinity,
		Status:          fireactions.RunnerStatus{State: fireactions.RunnerStatePending, Description: "Created"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return runner
}
