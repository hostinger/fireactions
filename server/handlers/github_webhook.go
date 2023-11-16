package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/config"
	"github.com/hostinger/fireactions/server/scheduler"
	"github.com/hostinger/fireactions/server/store"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	webhooks "github.com/go-playground/webhooks/v6/github"
)

// WebhookHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint POST /api/v1/github/webhook
func GitHubWebhookHandlerFunc(
	logger *zerolog.Logger, store store.Store, scheduler *scheduler.Scheduler, config *config.GitHubConfig,
) gin.HandlerFunc {
	hook, err := webhooks.New(webhooks.Options.Secret(config.WebhookSecret))
	if err != nil {
		panic(err)
	}

	jobHandler := newJobHandler(logger, scheduler, store, config)

	f := func(ctx *gin.Context) {
		data, err := hook.Parse(ctx.Request, webhooks.WorkflowJobEvent)
		if err != nil {
			ctx.JSON(500, gin.H{"message": fmt.Sprintf("error parsing GitHub webhook payload: %s", err.Error())})
			return
		}

		event, ok := data.(webhooks.WorkflowJobPayload)
		if !ok {
			logger.Warn().Msgf("skipping GitHub event: not a 'workflow_job' event. Make sure to only send 'workflow_job' events to this endpoint.")
			ctx.JSON(200, gin.H{
				"message": "Skipped event due to not being a workflow_job event"})
			return
		}

		err = jobHandler.Handle(ctx, &event)
		if err != nil {
			ctx.JSON(500, gin.H{"message": fmt.Sprintf("error handling job: %s", err.Error())})
			return
		}

		ctx.JSON(200, gin.H{"message": "OK"})
	}

	return f
}

type jobHandler struct {
	store     store.Store
	scheduler *scheduler.Scheduler
	config    *config.GitHubConfig
	logger    *zerolog.Logger
}

func newJobHandler(
	logger *zerolog.Logger, scheduler *scheduler.Scheduler, store store.Store, config *config.GitHubConfig) *jobHandler {
	jH := &jobHandler{
		scheduler: scheduler,
		store:     store,
		config:    config,
		logger:    logger,
	}

	return jH
}

func (h *jobHandler) Handle(ctx context.Context, j *webhooks.WorkflowJobPayload) error {
	logger := h.logger.With().
		Str("organisation", j.Organization.Login).
		Logger()

	if !lo.Contains(j.WorkflowJob.Labels, "self-hosted") {
		logger.Debug().Msgf("skipped job %d: not using self-hosted label", j.WorkflowJob.ID)
		return nil
	}

	l, found := lo.Find(j.WorkflowJob.Labels, func(label string) bool {
		return strings.HasPrefix(label, h.config.JobLabelPrefix)
	})

	if !found {
		logger.Debug().Msgf("skipped job %d: not using fireactions label", j.WorkflowJob.ID)
		return nil
	}

	jobLabel := strings.TrimPrefix(l, h.config.JobLabelPrefix)
	jobLabelConfig, ok := h.config.GetJobLabelConfig(jobLabel)
	if !ok {
		logger.Debug().Msgf("skipped job %d: no config for label %s", j.WorkflowJob.ID, jobLabel)
		return nil
	}

	if !lo.ContainsBy(jobLabelConfig.AllowedRepositories, func(item string) bool {
		regexp, err := regexp.Compile(item)
		if err != nil {
			return false
		}

		return regexp.MatchString(j.Repository.FullName)
	}) {
		logger.Debug().Msgf("skipped job %d: repository %s is not allowed to use label %s", j.WorkflowJob.ID, j.Repository.FullName, jobLabel)
		return nil
	}

	switch j.WorkflowJob.Status {
	case "queued":
		err := h.handleQueued(ctx, j, jobLabelConfig)
		if err != nil {
			return fmt.Errorf("queued: %w", err)
		}

		logger.Info().Msgf("created GitHub runner for job %d", j.WorkflowJob.ID)
	case "in_progress":
		err := h.handleInProgress(ctx, j)
		if err != nil {
			return fmt.Errorf("in_progress: %w", err)
		}
	case "completed":
		err := h.handleCompleted(ctx, j)
		if err != nil {
			return fmt.Errorf("completed: %w", err)
		}

		logger.Info().Msgf("deleted GitHub runner for job %d", j.WorkflowJob.ID)
	}

	return nil
}

func (h *jobHandler) handleQueued(ctx context.Context, j *webhooks.WorkflowJobPayload, jobLabelConfig *config.GitHubJobLabelConfig) error {
	runner := newRunnerFromJobPayload(j, jobLabelConfig)

	if err := h.store.CreateRunner(ctx, runner); err != nil {
		return fmt.Errorf("error creating runner: %w", err)
	}

	h.scheduler.AddToQueue(runner)
	return nil
}

func (h *jobHandler) handleInProgress(ctx context.Context, j *webhooks.WorkflowJobPayload) error {
	runner, err := h.store.GetRunnerByName(ctx, j.WorkflowJob.RunnerName)
	switch err {
	case nil, store.ErrNotFound:
		break
	default:
		return fmt.Errorf("store: getting runner: %w", err)
	}

	_, err = h.store.SetRunnerStatus(ctx, runner.ID, fireactions.RunnerStatus{Phase: fireactions.RunnerPhaseActive})
	if err != nil {
		return fmt.Errorf("store: setting runner status: %w", err)
	}

	return nil
}

func (h *jobHandler) handleCompleted(ctx context.Context, j *webhooks.WorkflowJobPayload) error {
	runner, err := h.store.GetRunnerByName(ctx, j.WorkflowJob.RunnerName)
	switch err {
	case nil, store.ErrNotFound:
		break
	default:
		return fmt.Errorf("store: getting runner: %w", err)
	}

	_, err = h.store.SetRunnerStatus(ctx, runner.ID, fireactions.RunnerStatus{Phase: fireactions.RunnerPhaseCompleted})
	if err != nil {
		return fmt.Errorf("store: setting runner status: %w", err)
	}

	return nil
}

func newRunnerFromJobPayload(j *webhooks.WorkflowJobPayload, jobLabelConfig *config.GitHubJobLabelConfig) *fireactions.Runner {
	runnerID := uuid.New().String()
	runner := &fireactions.Runner{
		ID:           runnerID,
		Name:         fmt.Sprintf("fireactions-%s", runnerID),
		NodeID:       nil,
		Image:        jobLabelConfig.Runner.Image,
		Status:       fireactions.RunnerStatus{Phase: fireactions.RunnerPhasePending},
		Organisation: j.Organization.Login,
		Labels:       j.WorkflowJob.Labels,
		Resources:    fireactions.RunnerResources{VCPUs: jobLabelConfig.Runner.Resources.VCPUs, MemoryBytes: jobLabelConfig.Runner.Resources.MemoryMB * 1024 * 1024},
		CreatedAt:    j.WorkflowJob.StartedAt,
		UpdatedAt:    j.WorkflowJob.StartedAt,
		DeletedAt:    nil,
	}

	if jobLabelConfig.Runner.ImagePullPolicy == "" {
		runner.ImagePullPolicy = fireactions.RunnerImagePullPolicyIfNotPresent
	} else {
		runner.ImagePullPolicy = fireactions.RunnerImagePullPolicy(jobLabelConfig.Runner.ImagePullPolicy)
	}

	for _, expression := range jobLabelConfig.Runner.Affinity {
		affinity := &fireactions.RunnerAffinityExpression{Key: expression.Key, Operator: expression.Operator, Values: expression.Values}
		runner.Affinity = append(runner.Affinity, affinity)
	}

	return runner
}
