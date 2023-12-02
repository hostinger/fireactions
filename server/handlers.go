package server

import (
	"net/http"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/gin-gonic/gin"
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
