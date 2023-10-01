package handler

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
)

// RegisterJobsV1 registers all HTTP handlers for the Jobs v1 API.
func RegisterJobsV1(r gin.IRouter, log *zerolog.Logger, store store.Store) {
	r.DELETE("/jobs/:id",
		DeleteJobHandlerFuncV1(log, store))
	r.GET("/jobs",
		GetJobsHandlerFuncV1(log, store))
	r.GET("/jobs/:id",
		GetJobHandlerFuncV1(log, store))
}

// GetJobsHandlerFuncV1 returns a HTTP handler function that returns all Jobs. The Jobs are returned in the v1
// format.
func GetJobsHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		type query struct {
			Organisation string `form:"organisation" binding:"-"`
			Repository   string `form:"repository" binding:"-"`
		}

		var q query
		ctx.ShouldBindQuery(&q)

		jobs, err := store.ListJobs(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		jobs = structs.FilterJobs(jobs, func(job *structs.Job) bool {
			if q.Organisation != "" && job.Organisation != q.Organisation {
				return false
			}

			if q.Repository != "" && job.Repository != q.Repository {
				return false
			}

			return true
		})

		ctx.JSON(200, gin.H{"jobs": convertJobsToJobsV1(jobs...)})
	}

	return f
}

// GetJobHandlerFuncV1 returns a HTTP handler function that returns a single Job by ID. The Job is returned in
// the v1 format.
func GetJobHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(c *gin.Context) {
		id := c.Param("id")

		job, err := store.GetJob(c, id)
		if err != nil {
			httperr.E(c, err)
			return
		}

		c.JSON(200, convertJobtoJobV1(job))
	}

	return f
}

// DeleteJobHandlerFuncV1 returns a HTTP handler function that deletes a single Job by ID.
func DeleteJobHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(c *gin.Context) {
		id := c.Param("id")

		if err := store.DeleteJob(c, id); err != nil {
			httperr.E(c, err)
			return
		}

		c.Status(204)
	}

	return f
}

func convertJobtoJobV1(job *structs.Job) *v1.Job {
	j := &v1.Job{
		ID:           job.ID,
		Name:         job.Name,
		Organisation: job.Organisation,
		Repository:   job.Repository,
		Status:       string(job.Status),
		CreatedAt:    job.CreatedAt,
	}

	return j
}

func convertJobsToJobsV1(jobs ...*structs.Job) []*v1.Job {
	j := make([]*v1.Job, 0, len(jobs))
	for _, job := range jobs {
		j = append(j, convertJobtoJobV1(job))
	}

	return j
}
