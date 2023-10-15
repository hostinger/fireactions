package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hostinger/fireactions/server/ghlabel"
	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	webhooks "github.com/go-playground/webhooks/v6/github"
)

// Scheduler is an interface that enqueues a Runner for scheduling to the best-fitting Node.
type Scheduler interface {
	Schedule(runner *models.Runner) error
	HandleEvent(event *models.Event)
}

func GetGitHubWebhookHandlerFuncV1(log *zerolog.Logger,
	secretKey string, jobLabelPrefix string, scheduler Scheduler, storer store.Store) gin.HandlerFunc {
	hook, err := webhooks.New(webhooks.Options.Secret(secretKey))
	if err != nil {
		log.Err(err).Msg("error creating GitHub webhook handlers")
		f := func(ctx *gin.Context) {
			ctx.JSON(500, gin.H{"message": "error creating GitHub webhook handlers"})
		}

		return f
	}

	f := func(ctx *gin.Context) {
		data, err := hook.Parse(ctx.Request, webhooks.WorkflowJobEvent)
		if err != nil {
			log.Err(err).Msg("error parsing GitHub webhook payload")
			ctx.JSON(500, gin.H{
				"message": fmt.Sprintf("error parsing GitHub webhook payload: %s", err.Error())})
			return
		}

		event, ok := data.(webhooks.WorkflowJobPayload)
		if !ok {
			log.Debug().Msg("skipping GitHub event: not a 'workflow_job' event. Make sure to only send 'workflow_job' events to this endpoint.")
			ctx.JSON(200, gin.H{
				"message": "Skipped event due to not being a workflow_job event"})
			return
		}

		jobID := fmt.Sprintf("%d", event.WorkflowJob.ID)

		if !lo.Contains(event.WorkflowJob.Labels, "self-hosted") {
			log.Debug().Msgf("skipped job %s: not using self-hosted label", jobID)
			ctx.JSON(200, gin.H{
				"message": "Skipped job due to not using 'self-hosted' label"})
			return
		}

		l, found := lo.Find(event.WorkflowJob.Labels, func(str string) bool {
			return strings.HasPrefix(str, jobLabelPrefix)
		})

		if !found {
			log.Debug().Msgf("skipped job %s: not using label with prefix '%s'", jobID, jobLabelPrefix)
			ctx.JSON(200, gin.H{
				"message": "Skipped job due to not using label with prefix"})
			return
		}

		label, err := ghlabel.New(l)
		if err != nil {
			log.Debug().Msgf("skipped job %s: error parsing label '%s': %s", jobID, l, err.Error())
			ctx.JSON(200, gin.H{
				"message": fmt.Sprintf("Skipped job due to error parsing label: %s", err.Error()),
			})
			return
		}

		group, err := findGroupByLabel(ctx, storer, label)
		if err != nil {
			log.Debug().Msgf("skipped job %s: error getting group %s: %s", jobID, label.Group, err.Error())
			ctx.JSON(200, gin.H{
				"message": fmt.Sprintf("Skipped job due to error getting group: %s", err.Error()),
			})
			return
		}

		flavor, err := findFlavorByLabel(ctx, storer, label)
		if err != nil {
			log.Debug().Msgf("skipped job %s: error getting flavor %s: %s", jobID, label.Flavor, err.Error())
			ctx.JSON(200, gin.H{
				"message": fmt.Sprintf("Skipped job due to error getting flavor: %s", err.Error()),
			})
			return
		}

		switch event.WorkflowJob.Status {
		case "queued":
			err := storer.SaveJob(ctx, &models.Job{
				ID:           jobID,
				Organisation: event.Organization.Login,
				Status:       models.JobStatusQueued,
				Name:         event.WorkflowJob.Name,
				Repository:   event.Repository.FullName,
				CreatedAt:    time.Now(),
			})
			if err != nil {
				log.Err(err).Msgf("error saving job %s", jobID)
				ctx.JSON(500, gin.H{
					"message": "Something went wrong, check the logs of Fireactions server for more information"})
				return
			}

			log.Debug().Msgf("created job %s", jobID)

			id := uuid.New().String()
			r := &models.Runner{
				ID:           id,
				Name:         fmt.Sprintf("runner-%s", id),
				Organisation: event.Organization.Login,
				Status:       models.RunnerStatusPending,
				Labels:       strings.Join(event.WorkflowJob.Labels, ","),
				Flavor:       flavor,
				Group:        group,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			err = storer.SaveRunner(ctx, r)
			if err != nil {
				log.Err(err).Msgf("error creating runner for job %s", jobID)
				ctx.JSON(500, gin.H{
					"message": "Something went wrong, check the logs of Fireactions server for more information"})
				return
			}

			log.Debug().Msgf("created runner %s for job %s", r.ID, jobID)

			err = scheduler.Schedule(r)
			if err != nil {
				log.Err(err).Msgf("error scheduling runner %s for job %s", id, jobID)
				ctx.JSON(500, gin.H{
					"message": "Something went wrong, check the logs of Fireactions server for more information"})
				return
			}
		case "in_progress":
			job, err := storer.GetJob(ctx, jobID)
			if err != nil {
				log.Err(err).Msgf("error getting job %s", jobID)
				ctx.JSON(500, gin.H{
					"message": "Something went wrong, check the logs of Fireactions server for more information"})
				return
			}

			job.Status = models.JobStatusInProgress
			err = storer.SaveJob(ctx, job)
			if err != nil {
				log.Err(err).Msgf("error updating job %s", jobID)
				ctx.JSON(500, gin.H{
					"message": "Something went wrong, check the logs of Fireactions server for more information"})
				return
			}

			log.Debug().Msgf("updated job %s to in progress", jobID)
		case "completed":
			job, err := storer.GetJob(ctx, jobID)
			if err != nil {
				log.Err(err).Msgf("error getting job %s", jobID)
				ctx.JSON(500, gin.H{
					"message": "Something went wrong, check the logs of Fireactions server for more information"})
				return
			}

			err = storer.DeleteJob(ctx, job.ID)
			if err != nil {
				log.Err(err).Msgf("error deleting job %s", jobID)
				ctx.JSON(500, gin.H{
					"message": "Something went wrong, check the logs of Fireactions server for more information"})
				return
			}

			log.Debug().Msgf("marked job %s as complete and deleted", jobID)
		default:
			log.Trace().Msgf("skipped job %s: unknown status of workflow_job event %s", jobID, event.WorkflowJob.Status)
		}

		ctx.JSON(200, gin.H{"message": "OK"})
	}

	return f
}

func findFlavorByLabel(ctx context.Context, store store.Store, label *ghlabel.Label) (*models.Flavor, error) {
	var flavor *models.Flavor
	var err error
	if label.Flavor == "" {
		flavor, err = store.GetDefaultFlavor(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		flavor, err = store.GetFlavor(ctx, label.Flavor)
		if err != nil {
			return nil, err
		}

		if !flavor.Enabled {
			return nil, errors.New("flavor is disabled")
		}
	}

	return flavor, nil
}

func findGroupByLabel(ctx context.Context, store store.Store, label *ghlabel.Label) (*models.Group, error) {
	var group *models.Group
	var err error
	if label.Group == "" {
		group, err = store.GetDefaultGroup(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		group, err = store.GetGroup(ctx, label.Group)
		if err != nil {
			return nil, err
		}

		if !group.Enabled {
			return nil, errors.New("group is disabled")
		}
	}

	return group, nil
}
