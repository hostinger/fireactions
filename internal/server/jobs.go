package server

import (
	"github.com/gin-gonic/gin"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/structs"
)

func (s *Server) handleGetJob(ctx *gin.Context) {
	id := ctx.Param("id")

	job, err := s.Store.GetJob(ctx, id)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, convertJobtoJobV1(job))
}

func (s *Server) handleGetJobs(ctx *gin.Context) {
	type query struct {
		Organisation string `form:"organisation"`
	}

	var q query
	err := ctx.BindQuery(&q)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	jobs, err := s.Store.GetJobs(ctx)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	jobs = jobs.Filter(func(job *structs.Job) bool {
		if q.Organisation != "" && job.Organisation != q.Organisation {
			return false
		}

		return true
	})

	ctx.JSON(200, gin.H{"jobs": convertJobsToJobsV1(jobs)})
}

func (s *Server) handleDelJob(ctx *gin.Context) {
	err := s.Store.DeleteJob(ctx, ctx.Param("id"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "Job deleted successfully"})
}

func convertJobtoJobV1(job *structs.Job) *api.Job {
	j := &api.Job{
		ID:           job.ID,
		Name:         job.Name,
		Organisation: job.Organisation,
		Repository:   job.Repository,
		Status:       string(job.Status),
		CompletedAt:  job.CompletedAt,
		CreatedAt:    job.CreatedAt,
	}

	return j
}

func convertJobsToJobsV1(jobs []*structs.Job) []*api.Job {
	j := make([]*api.Job, 0, len(jobs))
	for _, job := range jobs {
		j = append(j, convertJobtoJobV1(job))
	}

	return j
}
