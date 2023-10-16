package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1 "github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/server/httperr"
	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store"
	"github.com/rs/zerolog"
)

func RegisterNodesV1(r gin.IRouter, log *zerolog.Logger, scheduler Scheduler, store store.Store) {
	r.GET("/nodes",
		GetNodesHandlerFuncV1(log, store))
	r.GET("/nodes/:id",
		GetNodeHandlerFuncV1(log, store))
	r.POST("/nodes",
		RegisterNodeHandlerFuncV1(log, scheduler, store))
	r.DELETE("/nodes/:id",
		DeregisterNodeHandlerFuncV1(log, scheduler, store))
	r.POST("/nodes/:id/connect",
		ConnectNodeHandlerFuncV1(log, scheduler, store))
	r.POST("/nodes/:id/disconnect",
		DisconnectNodeHandlerFuncV1(log, scheduler, store))
	r.GET("/nodes/:id/runners",
		GetNodeRunnersHandlerFuncV1(log, scheduler, store))
	r.POST("/nodes/:id/runners/:runner/complete",
		CompleteNodeRunnerHandlerFuncV1(log, store))
	r.POST("/nodes/:id/runners/:runner/accept",
		AcceptNodeRunnerHandlerFuncV1(log, store))
	r.POST("/nodes/:id/runners/:runner/reject",
		RejectNodeRunnerHandlerFuncV1(log, store))
	r.POST("/nodes/:id/cordon",
		CordonNodeHandlerFuncV1(log, scheduler, store))
	r.POST("/nodes/:id/uncordon",
		UncordonNodeHandlerFuncV1(log, scheduler, store))
}

func GetNodesHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		type query struct {
			Organisation string `form:"organisation" binding:"-"`
			Group        string `form:"group" binding:"-"`
		}

		var q query
		ctx.ShouldBindQuery(&q)

		nodes, err := store.ListNodes(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		nodes = models.FilterNodes(nodes, func(node *models.Node) bool {
			if q.Organisation != "" && node.Organisation != q.Organisation {
				return false
			}

			if q.Group != "" {
				for _, group := range node.Groups {
					if group.Name == q.Group {
						return true
					}
				}

				return false
			}

			return true
		})

		ctx.JSON(200, gin.H{"nodes": convertNodesToNodesV1(nodes...)})
	}

	return f
}

func GetNodeHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")

		node, err := store.GetNode(ctx, id)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, convertNodeToNodeV1(node))
	}

	return f
}

func RegisterNodeHandlerFuncV1(log *zerolog.Logger, scheduler Scheduler, storer store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		var req v1.NodeRegisterRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			return
		}

		var groups []*models.Group
		for _, group := range req.Groups {
			g, err := storer.GetGroup(ctx, group)
			switch err.(type) {
			case nil:
				break
			case *store.ErrNotFound:
				ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("group %s doesn't exist", group)})
				return
			default:
				ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}

			groups = append(groups, g)
		}

		_, err = storer.GetNodeByName(ctx, req.Hostname)
		switch err.(type) {
		case nil:
			ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("node %s already exists", req.Hostname)})
			return
		case *store.ErrNotFound:
			break
		default:
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}

		uuid := uuid.New().String()
		n := &models.Node{
			ID:           uuid,
			Name:         req.Hostname,
			Organisation: req.Organisation,
			Groups:       groups,
			Status:       models.NodeStatusUnknown,
			CPU:          models.Resource{Capacity: int64(req.CpuTotal), Allocated: 0, OvercommitRatio: req.CpuOvercommitRatio},
			RAM:          models.Resource{Capacity: int64(req.MemTotal), Allocated: 0, OvercommitRatio: req.MemOvercommitRatio},
			IsCordoned:   true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = storer.SaveNode(ctx, n)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		log.Info().Msgf("registered new Node: %s", n)
		scheduler.HandleEvent(models.NewNodeEvent(models.EventTypeNodeCreated, n))

		ctx.JSON(200, &v1.NodeRegistrationInfo{ID: uuid})
	}

	return f
}

func DeregisterNodeHandlerFuncV1(log *zerolog.Logger, scheduler Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")

		n, err := store.GetNode(ctx, id)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		err = store.DeleteNode(ctx, id)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		scheduler.HandleEvent(models.NewNodeEvent(models.EventTypeNodeDeleted, n))
		log.Info().Msgf("deregistered Node: %s", n)
		ctx.Status(204)
	}

	return f
}

func ConnectNodeHandlerFuncV1(log *zerolog.Logger, scheduler Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")

		n, err := store.GetNode(ctx, id)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		n.Status = models.NodeStatusOnline
		err = store.SaveNode(ctx, n)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		runners, err := store.ListRunners(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		runners = models.FilterRunners(runners, func(runner *models.Runner) bool {
			if runner.Node == nil {
				return false
			}

			return runner.Node.ID == n.ID
		})

		// TODO: We should probably do this in a separate goroutine. Mark the runners as terminating and then delete them.
		for _, runner := range runners {
			err := store.DeleteRunner(ctx, runner.ID)
			if err != nil {
				httperr.E(ctx, err)
				return
			}

			log.Info().Msgf("deleted runner %s due to node %s reconnecting", runner.Name, n.Name)
		}

		scheduler.HandleEvent(models.NewNodeEvent(models.EventTypeNodeUpdated, n))
		log.Info().Msgf("updated Node status: %s", n)
		ctx.Status(204)
	}

	return f
}

func DisconnectNodeHandlerFuncV1(log *zerolog.Logger, scheduler Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")

		n, err := store.GetNode(ctx, id)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		n.Status = models.NodeStatusOffline
		err = store.SaveNode(ctx, n)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		runners, err := store.ListRunners(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		runners = models.FilterRunners(runners, func(runner *models.Runner) bool {
			if runner.Node == nil {
				return false
			}

			return runner.Node.ID == n.ID
		})

		// TODO: We should probably do this in a separate goroutine. Mark the runners as terminating and then delete them.
		for _, runner := range runners {
			err := store.DeleteRunner(ctx, runner.ID)
			if err != nil {
				httperr.E(ctx, err)
				return
			}

			log.Info().Msgf("deleted runner %s due to node %s disconnecting", runner.Name, n.Name)
		}

		scheduler.HandleEvent(models.NewNodeEvent(models.EventTypeNodeUpdated, n))
		log.Info().Msgf("updated Node status: %s", n)
		ctx.Status(204)
	}

	return f
}

func AcceptNodeRunnerHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		n, err := store.GetNode(ctx, ctx.Param("id"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		r, err := store.GetRunner(ctx, ctx.Param("runner"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		r.Status = models.RunnerStatusAccepted
		err = store.SaveRunner(ctx, r)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		log.Info().Msgf("runner %s accepted by node %s", r.Name, n.Name)
		ctx.Status(204)
	}

	return f
}

func RejectNodeRunnerHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		n, err := store.GetNode(ctx, ctx.Param("id"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		r, err := store.GetRunner(ctx, ctx.Param("runner"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		r.Status = models.RunnerStatusRejected
		err = store.SaveRunner(ctx, r)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		log.Info().Msgf("runner %s rejected by node %s", r.Name, n.Name)
		ctx.Status(204)
	}

	return f
}

func CompleteNodeRunnerHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		n, err := store.GetNode(ctx, ctx.Param("id"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		r, err := store.GetRunner(ctx, ctx.Param("runner"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		err = store.DeleteRunner(ctx, r.ID)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		log.Info().Msgf("runner %s completed on node %s", r.Name, n.Name)
		ctx.Status(204)
	}

	return f
}

func GetNodeRunnersHandlerFuncV1(log *zerolog.Logger, scheduler Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")

		node, err := store.GetNode(ctx, id)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		runners, err := store.ListRunners(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		runners = models.FilterRunners(runners, func(runner *models.Runner) bool {
			if runner.Node == nil || runner.Status != models.RunnerStatusAssigned {
				return false
			}

			return runner.Node.Name == node.Name
		})

		node.UpdatedAt = time.Now()
		err = store.SaveNode(ctx, node)
		if err != nil {
			httperr.E(ctx, err)
			return
		}
		scheduler.HandleEvent(models.NewNodeEvent(models.EventTypeNodeUpdated, node))

		ctx.JSON(200, gin.H{"runners": convertRunnersToRunnersV1(runners...)})
	}

	return f
}

func CordonNodeHandlerFuncV1(log *zerolog.Logger, scheduler Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		node, err := store.GetNode(ctx, ctx.Param("id"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		node.IsCordoned = true
		err = store.SaveNode(ctx, node)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		scheduler.HandleEvent(models.NewNodeEvent(models.EventTypeNodeUpdated, node))

		log.Info().Msgf("cordoned node %s", node.Name)
		ctx.Status(204)
	}

	return f
}

func UncordonNodeHandlerFuncV1(log *zerolog.Logger, scheduler Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		node, err := store.GetNode(ctx, ctx.Param("id"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		node.IsCordoned = false
		err = store.SaveNode(ctx, node)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		scheduler.HandleEvent(models.NewNodeEvent(models.EventTypeNodeUpdated, node))

		log.Info().Msgf("uncordoned node %s", node.Name)
		ctx.Status(204)
	}

	return f
}

func convertNodeToNodeV1(node *models.Node) *v1.Node {
	groups := make([]string, 0, len(node.Groups))
	for _, group := range node.Groups {
		groups = append(groups, group.Name)
	}

	n := &v1.Node{
		ID:           node.ID,
		Name:         node.Name,
		Organisation: node.Organisation,
		Groups:       groups,
		Status:       string(node.Status),
		CpuTotal:     node.CPU.Capacity,
		CpuFree:      node.CPU.Available(),
		MemTotal:     node.RAM.Capacity,
		MemFree:      node.RAM.Available(),
		IsCordoned:   node.IsCordoned,
		LastSeen:     node.UpdatedAt,
	}

	return n
}

func convertNodesToNodesV1(nodes ...*models.Node) v1.Nodes {
	n := make([]*v1.Node, 0, len(nodes))
	for _, node := range nodes {
		n = append(n, convertNodeToNodeV1(node))
	}

	return n
}
