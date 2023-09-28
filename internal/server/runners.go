package server

import (
	"github.com/gin-gonic/gin"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/structs"
)

func (s *Server) handleGetRunners(ctx *gin.Context) {
	type query struct {
		Organisation string `form:"organisation"`
		Group        string `form:"group"`
		Node         string `form:"node"`
	}

	var q query
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	runners, err := s.Store.GetRunners(ctx)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	runners = runners.Filter(func(runner *structs.Runner) bool {
		if q.Organisation != "" && runner.Organisation != q.Organisation {
			return false
		}

		if q.Group != "" && runner.Group.Name != q.Group {
			return false
		}

		if q.Node != "" && runner.Node != nil && runner.Node.Name != q.Node {
			return false
		}

		return true
	})

	ctx.JSON(200, gin.H{"runners": convertRunnersToRunnersV1(runners...)})
}

func (s *Server) handleGetRunner(ctx *gin.Context) {
	runner, err := s.Store.GetRunner(ctx, ctx.Param("id"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, convertRunnerToRunnerV1(runner))
}

func convertRunnersToRunnersV1(runner ...*structs.Runner) api.Runners {
	runners := make([]*api.Runner, 0, len(runner))
	for _, r := range runner {
		runners = append(runners, convertRunnerToRunnerV1(r))
	}

	return runners
}

func convertRunnerToRunnerV1(runner *structs.Runner) *api.Runner {
	r := &api.Runner{
		ID:           runner.ID,
		Organisation: runner.Organisation,
		Group:        convertGroupToGroupV1(runner.Group),
		Node:         convertNodeToNodeV1(runner.Node),
		Name:         runner.Name,
		Status:       string(runner.Status),
		Flavor:       convertFlavorToFlavorV1(runner.Flavor),
		Labels:       runner.Labels,
		CreatedAt:    runner.CreatedAt,
		UpdatedAt:    runner.UpdatedAt,
	}

	return r
}
