package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	webhooks "github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/hostinger/fireactions/internal/server/ghlabel"
	"github.com/hostinger/fireactions/internal/structs"
	"github.com/samber/lo"
)

func (s *Server) handleGitHubRegistrationToken(ctx *gin.Context) {
	org := ctx.Param("organisation")

	token, err := s.ghClient.GetRegistrationToken(ctx, org)
	if err != nil {
		s.log.Err(err).Msgf("error getting GitHub registration token for %s", org)
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("error getting GitHub registration token: %s", err.Error())})
		return
	}

	ctx.JSON(200, gin.H{"token": token})
}

func (s *Server) handleGitHubWebhook(ctx *gin.Context) {
	hook, err := webhooks.New(webhooks.Options.Secret(s.cfg.GitHub.WebhookSecret))
	if err != nil {
		s.log.Err(err).Msg("error creating GitHub webhook handlers")
		ctx.JSON(500, gin.H{"message": fmt.Sprintf("error creating GitHub webhook handlers: %s", err.Error())})
		return
	}

	data, err := hook.Parse(ctx.Request, webhooks.WorkflowJobEvent)
	if err != nil {
		s.log.Err(err).Msg("error parsing GitHub webhook payload")
		ctx.JSON(500, gin.H{"message": fmt.Sprintf("error parsing GitHub webhook payload: %s", err.Error())})
		return
	}

	event, ok := data.(webhooks.WorkflowJobPayload)
	if !ok {
		s.log.Debug().Msg("skipping GitHub event: not a 'workflow_job' event. Make sure to only send 'workflow_job' events to this endpoint.")
		ctx.JSON(200, gin.H{"message": "Skipped event due to not being a workflow_job event"})
		return
	}

	jobID := fmt.Sprintf("%d", event.WorkflowJob.ID)

	if !lo.Contains(event.WorkflowJob.Labels, "self-hosted") {
		s.log.Debug().Msgf("skipped job %s: not using self-hosted label", jobID)
		ctx.JSON(200, gin.H{"message": "Skipped job due to not using 'self-hosted' label"})
		return
	}

	l, found := lo.Find(event.WorkflowJob.Labels, func(str string) bool {
		return strings.HasPrefix(str, s.cfg.GitHub.JobLabelPrefix)
	})

	if !found {
		s.log.Debug().Msgf("skipped job %s: not using label with prefix '%s'", jobID, s.cfg.GitHub.JobLabelPrefix)
		ctx.JSON(200, gin.H{"message": "Skipped job due to not using label with prefix"})
		return
	}

	label, err := ghlabel.New(strings.TrimPrefix(l, s.cfg.GitHub.JobLabelPrefix), ghlabel.WithDefaultFlavor(s.cfg.DefaultFlavor))
	if err != nil {
		s.log.Debug().Msgf("skipped job %s: unrecognized label: %s", jobID, err.Error())
		ctx.JSON(200, gin.H{"message": fmt.Sprintf("Skipped job due to unrecognized label: %s", err.Error())})
		return
	}

	flavor, err := s.fm.GetFlavor(label.Flavor)
	if err != nil {
		s.log.Debug().Msgf("skipped job %s: unrecognized flavor %s: %s", jobID, label.Flavor, err.Error())
		ctx.JSON(200, gin.H{"message": fmt.Sprintf("Skipped job due to unrecognized flavor: %s", err.Error())})
		return
	}

	switch event.WorkflowJob.Status {
	case "queued":
		err = s.Store.CreateJob(ctx, &structs.Job{
			ID:           jobID,
			Organisation: event.Organization.Login,
			Status:       structs.JobStatusQueued,
			Name:         event.WorkflowJob.Name,
			Repository:   event.Repository.FullName,
			CreatedAt:    time.Now(),
		})
		if err != nil {
			s.log.Err(err).Msgf("error creating job %s", jobID)
			ctx.JSON(500, gin.H{"message": fmt.Sprintf("error creating job: %s", err.Error())})
			return
		}
		s.log.Info().Msgf("job %s created", jobID)

		id := uuid.New().String()
		r := &structs.Runner{
			ID:           id,
			Name:         fmt.Sprintf("runner-%s", id),
			Organisation: event.Organization.Login,
			Status:       structs.RunnerStatusPending,
			Labels:       strings.Join(event.WorkflowJob.Labels, ","),
			Flavor:       flavor,
			Group:        label.Group,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err = s.Store.CreateRunner(ctx, r)
		if err != nil {
			s.log.Err(err).Msgf("error creating runner for job %s", jobID)
			ctx.JSON(500, gin.H{"message": "Something went wrong, check the logs of Fireactions server for more information"})
			return
		}

		err = s.scheduler.Schedule(r)
		if err != nil {
			s.log.Err(err).Msgf("error scheduling runner %s for job %s", id, jobID)
			ctx.JSON(500, gin.H{"message": "Something went wrong, check the logs of Fireactions server for more information"})
			return
		}

		s.log.Info().Msgf("runner %s created for job %s", id, jobID)
	case "in_progress":
		job, err := s.Store.GetJob(ctx, jobID)
		if err != nil {
			s.log.Err(err).Msgf("error getting job %s", jobID)
			ctx.JSON(500, gin.H{"message": "Something went wrong, check the logs of Fireactions server for more information"})
			return
		}

		job.Status = structs.JobStatusInProgress
		err = s.Store.UpdateJob(ctx, job)
		if err != nil {
			s.log.Err(err).Msgf("error updating job %s", jobID)
			ctx.JSON(500, gin.H{"message": "Something went wrong, check the logs of Fireactions server for more information"})
			return
		}

		s.log.Info().Msgf("job %s updated to in progress", jobID)
	case "completed":
		job, err := s.Store.GetJob(ctx, jobID)
		if err != nil {
			s.log.Err(err).Msgf("error getting job %s", jobID)
			ctx.JSON(500, gin.H{"message": "Something went wrong, check the logs of Fireactions server for more information"})
			return
		}

		err = s.Store.DeleteJob(ctx, job.ID)
		if err != nil {
			s.log.Err(err).Msgf("error deleting job %s", jobID)
			ctx.JSON(500, gin.H{"message": "Something went wrong, check the logs of Fireactions server for more information"})
			return
		}

		s.log.Info().Msgf("job %s marked as complete and deleted", jobID)
	default:
		s.log.Trace().Msgf("skipped job %s: unknown status of workflow_job event %s", jobID, event.WorkflowJob.Status)
	}

	ctx.JSON(200, gin.H{"message": "OK"})
}
