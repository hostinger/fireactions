package server

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/pkg/stringid"
	"github.com/hostinger/fireactions/server/store"
	"github.com/hostinger/fireactions/version"
	"github.com/samber/lo"

	webhooks "github.com/go-playground/webhooks/v6/github"
)

func (s *Server) handleGitHubWebhook() gin.HandlerFunc {
	hook, err := webhooks.New(webhooks.Options.Secret(s.config.GitHubConfig.WebhookSecret))
	if err != nil {
		panic(err)
	}

	f := func(ctx *gin.Context) {
		data, err := hook.Parse(ctx.Request, webhooks.WorkflowJobEvent)
		if err != nil {
			ctx.JSON(400, fmt.Sprintf("Invalid webhook payload: %s. Make sure to send a JSON payload.", err.Error()))
			return
		}

		event, ok := data.(webhooks.WorkflowJobPayload)
		if !ok {
			s.logger.Warn().Msgf("skipping GitHub event: not a 'workflow_job' event. Make sure to only send 'workflow_job' events to this endpoint.")
			ctx.JSON(200, gin.H{
				"error": "Skipped event due to not being a workflow_job event"})
			return
		}

		err = s.handleGitHubWorkflowJob(ctx, &event)
		if err != nil {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("error handling job: %s", err.Error())})
			return
		}

		ctx.JSON(200, gin.H{"message": "OK"})
	}

	return f
}

func (s *Server) handleGitHubWorkflowJob(ctx context.Context, j *webhooks.WorkflowJobPayload) error {
	logger := s.logger.With().
		Str("organisation", j.Organization.Login).
		Logger()

	if !lo.Contains(j.WorkflowJob.Labels, "self-hosted") {
		logger.Debug().Msgf("skipped job %d: not using self-hosted label", j.WorkflowJob.ID)
		return nil
	}

	l, found := lo.Find(j.WorkflowJob.Labels, func(label string) bool {
		return strings.HasPrefix(label, s.config.GitHubConfig.JobLabelPrefix)
	})

	if !found {
		logger.Debug().Msgf("skipped job %d: not using fireactions label", j.WorkflowJob.ID)
		return nil
	}

	jobLabel := strings.TrimPrefix(l, s.config.GitHubConfig.JobLabelPrefix)
	jobLabelConfig, ok := s.config.GitHubConfig.GetJobLabelConfig(jobLabel)
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
		err := s.handleGitHubWorkflowJobQueued(ctx, j, jobLabelConfig)
		if err != nil {
			return fmt.Errorf("queued: %w", err)
		}
	case "in_progress":
		err := s.handleGitHubWorkflowJobInProgress(ctx, j)
		if err != nil {
			return fmt.Errorf("in_progress: %w", err)
		}
	case "completed":
		err := s.handleGitHubWorkflowJobCompleted(ctx, j)
		if err != nil {
			return fmt.Errorf("completed: %w", err)
		}
	}

	return nil
}

func (s *Server) handleGitHubWorkflowJobInProgress(ctx context.Context, j *webhooks.WorkflowJobPayload) error {
	logger := s.logger.With().
		Str("organisation", j.Organization.Login).
		Logger()

	runner, err := s.store.GetRunnerByName(ctx, j.WorkflowJob.RunnerName)
	if err != nil {
		if err == store.ErrNotFound {
			logger.Warn().Msgf("skipped updating GitHub runner %s for job %d: runner doesn't exist", j.WorkflowJob.RunnerName, j.WorkflowJob.ID)
			return nil
		}

		return fmt.Errorf("store: getting runner: %w", err)
	}

	if runner.Status.Phase == fireactions.RunnerPhaseActive {
		logger.Warn().Msgf("skipped updating GitHub runner %s for job %d: runner is already active (multiple jobs might have ran on this GitHub runner)",
			j.WorkflowJob.RunnerName, j.WorkflowJob.ID)
	}

	_, err = s.store.SetRunnerStatus(ctx, runner.ID, fireactions.RunnerStatus{Phase: fireactions.RunnerPhaseActive})
	if err != nil {
		return fmt.Errorf("store: setting runner status: %w", err)
	}

	logger.Info().Msgf("updated GitHub runner %s for job %d (job is in progress)", runner.Name, j.WorkflowJob.ID)
	return nil
}

func (s *Server) handleGitHubWorkflowJobCompleted(ctx context.Context, j *webhooks.WorkflowJobPayload) error {
	logger := s.logger.With().
		Str("organisation", j.Organization.Login).
		Logger()

	runner, err := s.store.GetRunnerByName(ctx, j.WorkflowJob.RunnerName)
	if err != nil {
		if err == store.ErrNotFound {
			logger.Warn().Msgf("skipped deleting GitHub runner %s for job %d: runner doesn't exist", j.WorkflowJob.RunnerName, j.WorkflowJob.ID)
			return nil
		}

		return fmt.Errorf("store: getting runner: %w", err)
	}

	if runner.Status.Phase == fireactions.RunnerPhaseCompleted {
		logger.Warn().Msgf("skipped deleting GitHub runner %s for job %d: runner is already completed (multiple jobs might have ran on this GitHub runner)",
			j.WorkflowJob.RunnerName, j.WorkflowJob.ID)
	}

	_, err = s.store.SetRunnerStatus(ctx, runner.ID, fireactions.RunnerStatus{Phase: fireactions.RunnerPhaseCompleted})
	if err != nil {
		return fmt.Errorf("store: setting runner status: %w", err)
	}

	logger.Info().Msgf("deleted GitHub runner %s for job %d (job is completed)", runner.Name, j.WorkflowJob.ID)
	return nil
}

func (s *Server) handleGitHubWorkflowJobQueued(ctx context.Context, j *webhooks.WorkflowJobPayload, jobLabelConfig *GitHubJobLabelConfig) error {
	logger := s.logger.With().
		Str("organisation", j.Organization.Login).
		Logger()

	runner := newRunnerFromJobPayload(j, jobLabelConfig)

	if err := s.store.CreateRunner(ctx, runner); err != nil {
		return fmt.Errorf("store: creating runner: %w", err)
	}

	logger.Info().Msgf("created GitHub runner %s for job %d", runner.Name, j.WorkflowJob.ID)
	s.scheduler.AddToQueue(runner)
	return nil
}

func (s *Server) handleCreateRunner(ctx *gin.Context) {
	var request fireactions.CreateRunnerRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	if request.Count < 1 {
		ctx.JSON(400, gin.H{"error": "Count must be greater than 0"})
		return
	}

	jobLabel, ok := s.config.GitHubConfig.GetJobLabelConfig(request.JobLabel)
	if !ok {
		ctx.JSON(404, gin.H{"error": fmt.Sprintf("Job label %s doesn't exist", request.JobLabel)})
		return
	}

	var runners []*fireactions.Runner
	for i := 0; i < request.Count; i++ {
		runnerID := stringid.New()
		runner := &fireactions.Runner{
			ID:              runnerID,
			Name:            fmt.Sprintf("fireactions-%s", runnerID),
			NodeID:          nil,
			Image:           jobLabel.Runner.Image,
			ImagePullPolicy: fireactions.RunnerImagePullPolicy(jobLabel.Runner.ImagePullPolicy),
			Status:          fireactions.RunnerStatus{Phase: fireactions.RunnerPhasePending},
			Organisation:    request.Organisation,
			Labels:          []string{"self-hosted", fmt.Sprintf("%s%s", s.config.GitHubConfig.JobLabelPrefix, jobLabel.Name)},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			DeletedAt:       nil,
		}

		runner.Resources = fireactions.RunnerResources{
			VCPUs:       jobLabel.Runner.Resources.VCPUs,
			MemoryBytes: jobLabel.Runner.Resources.MemoryMB * 1024 * 1024,
		}

		if jobLabel.Runner.ImagePullPolicy == "" {
			runner.ImagePullPolicy = fireactions.RunnerImagePullPolicyIfNotPresent
		} else {
			runner.ImagePullPolicy = fireactions.RunnerImagePullPolicy(jobLabel.Runner.ImagePullPolicy)
		}

		for _, affinity := range jobLabel.Runner.Affinity {
			runner.Affinity = append(runner.Affinity, &fireactions.RunnerAffinityExpression{Key: affinity.Key, Operator: affinity.Operator, Values: affinity.Values})
		}

		runners = append(runners, runner)
	}

	err = s.store.CreateRunners(ctx, runners)
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	s.scheduler.AddToQueue(runners...)
	ctx.JSON(200, gin.H{"runners": runners})
}

func (s *Server) handleGetRunner(ctx *gin.Context) {
	runner, err := s.store.GetRunner(ctx, ctx.Param("id"))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	ctx.JSON(200, runner)
}

func (s *Server) handleGetRunners(ctx *gin.Context) {
	type Query struct {
		Organisation *string `form:"organisation"`
	}

	var query Query
	err := ctx.ShouldBindQuery(&query)
	if err != nil {
		ctx.JSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	runners, err := s.store.GetRunners(ctx, func(r *fireactions.Runner) bool {
		if r.DeletedAt != nil {
			return false
		}

		if query.Organisation != nil && *query.Organisation != r.Organisation {
			return false
		}

		return true
	})
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	ctx.JSON(200, gin.H{"runners": runners})
}

func (s *Server) handleGetRunnerRegistrationToken(ctx *gin.Context) {
	runner, err := s.store.GetRunner(ctx, ctx.Param("id"))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	token, err := s.github.GetRegistrationToken(ctx, runner.Organisation)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(200, fireactions.RunnerRegistrationToken{Token: token})
}

func (s *Server) handleGetRunnerRemoveToken(ctx *gin.Context) {
	runner, err := s.store.GetRunner(ctx, ctx.Param("id"))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	token, err := s.github.GetRemoveToken(ctx, runner.Organisation)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(200, fireactions.RunnerRemoveToken{Token: token})
}

func (s *Server) handleSetRunnerStatus(ctx *gin.Context) {
	var request fireactions.SetRunnerStatusRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	_, err = s.store.SetRunnerStatus(ctx, ctx.Param("id"), fireactions.RunnerStatus(request))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	ctx.Status(204)
}

func (s *Server) handleDeleteRunner(ctx *gin.Context) {
	err := s.store.DeallocateRunner(ctx, ctx.Param("id"))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	err = s.store.SoftDeleteRunner(ctx, ctx.Param("id"))
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	ctx.Status(204)
}

func (s *Server) handleRegisterNode(ctx *gin.Context) {
	var req fireactions.NodeRegisterRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	node, err := s.store.GetNodeByName(ctx, req.Name)
	if err != nil && err != store.ErrNotFound {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	if node != nil {
		node.CPU.Capacity = req.CpuCapacity
		node.RAM.Capacity = req.RamCapacity
		node.CPU.OvercommitRatio = req.CpuOvercommitRatio
		node.RAM.OvercommitRatio = req.RamOvercommitRatio
		node.Labels = req.Labels
		node.PollInterval = req.PollInterval
		node.UpdatedAt = time.Now()

		err = s.store.SaveNode(ctx, node)
		if err != nil {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
			return
		}

		s.scheduler.NotifyNodeUpdated(node)
		s.logger.Info().Msgf("updated node: %s", node.Name)
		ctx.JSON(200, &fireactions.NodeRegistrationInfo{ID: node.ID})
		return
	}

	node = &fireactions.Node{
		ID:           stringid.New(),
		Name:         req.Name,
		CPU:          fireactions.NodeResource{Allocated: 0, Capacity: req.CpuCapacity, OvercommitRatio: req.CpuOvercommitRatio},
		RAM:          fireactions.NodeResource{Allocated: 0, Capacity: req.RamCapacity, OvercommitRatio: req.RamOvercommitRatio},
		Status:       fireactions.NodeStatusCordoned,
		Labels:       req.Labels,
		PollInterval: req.PollInterval,
		LastPoll:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = s.store.SaveNode(ctx, node)
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	s.scheduler.NotifyNodeCreated(node)
	s.logger.Info().Msgf("registered node: %s", node.Name)
	ctx.JSON(200, &fireactions.NodeRegistrationInfo{ID: node.ID})
}

func (s *Server) handleCordonNode(ctx *gin.Context) {
	node, err := s.store.SetNodeStatus(ctx, ctx.Param("id"), fireactions.NodeStatusCordoned)
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	s.scheduler.NotifyNodeUpdated(node)
	s.logger.Info().Msgf("cordoned node: %s", node.Name)

	ctx.Status(204)
}

func (s *Server) handleUncordonNode(ctx *gin.Context) {
	node, err := s.store.SetNodeStatus(ctx, ctx.Param("id"), fireactions.NodeStatusReady)
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	s.scheduler.NotifyNodeUpdated(node)
	s.logger.Info().Msgf("uncordoned node: %s", node.Name)

	ctx.Status(204)
}

func (s *Server) handleDeregisterNode(ctx *gin.Context) {
	id := ctx.Param("id")
	node, err := s.store.GetNode(ctx, id)
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", id)})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	err = s.store.DeleteNode(ctx, node.ID)
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	s.scheduler.NotifyNodeDeleted(node)
	s.logger.Info().Msgf("deregistered node: %s", node.Name)
	ctx.Status(204)
}

func (s *Server) handleGetNodeRunners(ctx *gin.Context) {
	node, err := s.store.SetNodeLastPoll(ctx, ctx.Param("id"), time.Now())
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	runners, err := s.store.GetRunners(ctx, func(r *fireactions.Runner) bool {
		if r.NodeID == nil || r.DeletedAt != nil {
			return false
		}

		return *r.NodeID == node.ID
	})
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	s.scheduler.NotifyNodeUpdated(node)
	ctx.JSON(200, gin.H{"runners": runners})
}

func (s *Server) handleGetNode(ctx *gin.Context) {
	node, err := s.store.GetNode(ctx, ctx.Param("id"))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.JSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	ctx.JSON(200, node)
}

func (s *Server) handleGetNodes(ctx *gin.Context) {
	nodes, err := s.store.GetNodes(ctx, nil)
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	ctx.JSON(200, gin.H{"nodes": nodes})
}

func (s *Server) handleGetHealthz(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

func (s *Server) handleGetVersion(ctx *gin.Context) {
	ctx.String(http.StatusOK, version.String())
}

func newRunnerFromJobPayload(j *webhooks.WorkflowJobPayload, jobLabelConfig *GitHubJobLabelConfig) *fireactions.Runner {
	runnerID := stringid.New()
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
