package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/version"
)

func (s *Server) handleGetHealthz(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

func (s *Server) handleGetVersion(ctx *gin.Context) {
	ctx.String(http.StatusOK, version.String())
}

func (s *Server) handleWebhook() gin.HandlerFunc {
	handler := githubevents.New(s.config.GitHub.WebhookSecret)
	handler.OnWorkflowRunEventAny(s.createOrUpdateWorkflowRun)
	handler.OnWorkflowJobEventAny(s.createOrUpdateWorkflowJob)
	handler.OnWorkflowJobEventInProgress(s.updateRunnerForWorkflowJob)
	handler.OnWorkflowJobEventQueued(s.createRunnerForWorkflowJob)
	handler.OnWorkflowJobEventCompleted(s.updateRunnerForWorkflowJob)

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

func (s *Server) handleGetWorkflowRunStats(ctx *gin.Context) {
	organisation := ctx.Param("organisation")

	var query fireactions.WorkflowRunStatsQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	if query.Start.IsZero() {
		query.Start = time.Now().Add(-time.Hour * 24 * 7)
	}

	if query.End.IsZero() {
		query.End = time.Now()
	}

	if query.Sort == "" {
		query.Sort = "total"
	}

	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	if query.Limit == 0 {
		query.Limit = 100
	}

	if err := query.Validate(); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	workflowRuns, err := s.store.GetWorkflowRuns(ctx, func(wr *github.WorkflowRun) bool {
		if wr.GetRepository().GetOwner().GetLogin() != organisation {
			return false
		}

		if wr.GetStatus() != "completed" {
			return false
		}

		if query.Repositories != "" {
			repositories := strings.Split(query.Repositories, ",")
			found := false
			for _, repository := range repositories {
				if repository != wr.GetRepository().GetName() {
					continue
				}

				found = true
			}

			if !found {
				return false
			}
		}

		return wr.GetRunStartedAt().After(query.Start) && wr.GetRunStartedAt().Before(query.End)
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	workflowRunStats := fireactions.WorkflowRunStats{Stats: make([]*fireactions.WorkflowRunStat, 0)}

	statsByRepository := make(map[string]*fireactions.WorkflowRunStat)
	for _, workflowRun := range workflowRuns {
		stats := statsByRepository[workflowRun.GetRepository().GetName()]
		if stats == nil {
			stats = &fireactions.WorkflowRunStat{Total: 0, TotalDuration: 0, Succeeded: 0, Failed: 0, Cancelled: 0, Repository: workflowRun.GetRepository().GetName()}
			statsByRepository[workflowRun.GetRepository().GetName()] = stats
		}

		switch workflowRun.GetConclusion() {
		case "cancelled":
			stats.Cancelled += 1
		case "success":
			stats.Succeeded += 1
		case "failure":
			stats.Failed += 1
		default:
		}

		stats.TotalDuration += workflowRun.GetUpdatedAt().Sub(workflowRun.GetRunStartedAt().Time)
		stats.Total += 1
	}

	for _, stat := range statsByRepository {
		workflowRunStats.Stats = append(workflowRunStats.Stats, stat)
	}

	err = workflowRunStats.Sort(query.Sort, query.SortOrder)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	if query.Limit > 0 && len(workflowRunStats.Stats) > query.Limit {
		workflowRunStats.Stats = workflowRunStats.Stats[:query.Limit]
	}

	ctx.JSON(200, workflowRunStats)
}
