package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	v1 "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/server/structs"
	"github.com/rs/zerolog"
)

// RunnerStorer is an interface for storing and retrieving Runners.
type RunnerStorer interface {
	ListRunners(ctx context.Context) ([]*structs.Runner, error)
	GetRunner(ctx context.Context, id string) (*structs.Runner, error)
	SaveRunner(ctx context.Context, runner *structs.Runner) error
	DeleteRunner(ctx context.Context, id string) error
}

// RegisterRunnersV1 registers all HTTP handlers for the Runners v1 API.
func RegisterRunnersV1(r gin.IRouter, log *zerolog.Logger, rs RunnerStorer) {
	r.GET("/runners",
		GetRunnersHandlerFuncV1(log, rs))
	r.GET("/runners/:id",
		GetRunnerHandlerFuncV1(log, rs))
}

// GetRunnersHandlerFuncV1 returns a HTTP handler function that returns all Runners. The Runners are returned in the v1
// format.
func GetRunnersHandlerFuncV1(log *zerolog.Logger, rs RunnerStorer) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		type query struct {
			Organisation string `form:"organisation" binding:"-"`
			Group        string `form:"group" binding:"-"`
		}

		var q query
		ctx.ShouldBindQuery(&q)

		runners, err := rs.ListRunners(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		runners = structs.FilterRunners(runners, func(runner *structs.Runner) bool {
			if q.Organisation != "" && runner.Organisation != q.Organisation {
				return false
			}

			if q.Group != "" && runner.Group.Name != q.Group {
				return false
			}

			return true
		})

		ctx.JSON(200, gin.H{"runners": convertRunnersToRunnersV1(runners...)})
	}

	return f
}

// GetRunnerHandlerFuncV1 returns a HTTP handler function that returns a single Runner by ID. The Runner is returned in
// the v1 format.
func GetRunnerHandlerFuncV1(log *zerolog.Logger, rs RunnerStorer) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")

		runner, err := rs.GetRunner(ctx, id)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, convertRunnerToRunnerV1(runner))
	}

	return f
}

func convertRunnersToRunnersV1(runner ...*structs.Runner) v1.Runners {
	runners := make([]*v1.Runner, 0, len(runner))
	for _, r := range runner {
		runners = append(runners, convertRunnerToRunnerV1(r))
	}

	return runners
}

func convertRunnerToRunnerV1(runner *structs.Runner) *v1.Runner {
	r := &v1.Runner{
		ID:           runner.ID,
		Organisation: runner.Organisation,
		Group:        convertGroupToGroupV1(runner.Group),
		Name:         runner.Name,
		Status:       string(runner.Status),
		Flavor:       convertFlavorToFlavorV1(runner.Flavor),
		Labels:       runner.Labels,
		CreatedAt:    runner.CreatedAt,
		UpdatedAt:    runner.UpdatedAt,
	}

	if runner.Node != nil {
		r.Node = convertNodeToNodeV1(runner.Node)
	}

	return r
}
