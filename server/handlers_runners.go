package server

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/store"
)

func (s *Server) handleCreateRunner(ctx *gin.Context) {
	var request fireactions.CreateRunnerRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	if request.Count < 1 {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Count must be greater than 0"})
		return
	}

	jobLabel, ok := s.config.FindJobLabel(request.JobLabel)
	if !ok {
		ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Job label %s doesn't exist", request.JobLabel)})
		return
	}

	installation, _, err := s.github.Apps.FindOrganizationInstallation(ctx, request.Organisation)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Failed to fetch GitHub App installation: %s", err.Error())})
		return
	}

	client := s.github.InstallationClient(installation.GetID())
	var runners []*fireactions.Runner
	for i := 0; i < request.Count; i++ {
		runner := convertJobLabelConfigToRunner(jobLabel, request.Organisation)
		runners = append(runners, runner)

		var config *github.JITRunnerConfig
		req := &github.GenerateJITConfigRequest{
			Name:          runner.Name,
			RunnerGroupID: 1,
			Labels:        runner.Labels,
		}

		config, _, err = client.Actions.GenerateOrgJITConfig(ctx, runner.Organisation, req)
		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Failed to generate JIT config: %s", err.Error())})
			return
		}

		runner.Metadata["fireactions"] = map[string]interface{}{
			"runner_id":         runner.ID,
			"runner_jit_config": config.GetEncodedJITConfig(),
		}

		err = s.store.SaveRunner(ctx, runner)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
			return
		}

		s.scheduler.AddToQueue(runner)
	}

	ctx.JSON(200, gin.H{"runners": runners})
}

func (s *Server) handleGetRunner(ctx *gin.Context) {
	runner, err := s.store.GetRunner(ctx, ctx.Param("id"))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	ctx.JSON(200, runner)
}

func (s *Server) handleGetRunners(ctx *gin.Context) {
	runners, err := s.store.GetRunners(ctx, func(r *fireactions.Runner) bool {
		return r.DeletedAt == nil
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	ctx.JSON(200, gin.H{"runners": runners})
}

func (s *Server) handleSetRunnerStatus(ctx *gin.Context) {
	var request fireactions.SetRunnerStatusRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	_, err = s.store.UpdateRunner(ctx, ctx.Param("id"), func(r *fireactions.Runner) error {
		r.Status = fireactions.RunnerStatus(request)
		return nil
	})
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	ctx.Status(204)
}

func (s *Server) handleDeleteRunner(ctx *gin.Context) {
	runnerID := ctx.Param("id")

	runner, err := s.store.GetRunner(ctx, runnerID)
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Runner with ID %s doesn't exist", runnerID)})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	if runner.Status.State != fireactions.RunnerStateCompleted {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Runner with ID %s is not completed", runnerID)})
		return
	}

	if runner.DeletedAt != nil {
		ctx.Status(204)
		return
	}

	_, err = s.store.UpdateNode(ctx, runner.GetNodeID(), func(n *fireactions.Node) error {
		n.CPU.Release(runner.Resources.VCPUs)
		n.RAM.Release(runner.Resources.MemoryMB * 1024 * 1024)

		return nil
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	_, err = s.store.UpdateRunner(ctx, runnerID, func(r *fireactions.Runner) error {
		now := time.Now()
		r.DeletedAt = &now

		return nil
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	ctx.Status(204)
}
