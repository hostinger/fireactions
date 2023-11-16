package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler"
	"github.com/hostinger/fireactions/server/store"
	"github.com/rs/zerolog"
)

// RegisterNodesHandlers registers HTTP handlers for /api/v1/nodes/* endpoints
// to the provided router.
func RegisterNodesHandlers(logger *zerolog.Logger, r gin.IRouter, scheduler *scheduler.Scheduler, store store.Store) {
	nodes := r.Group("/nodes")
	{
		nodes.GET("", NodesHandlerFunc(logger, store))
		nodes.GET("/:id", NodeHandlerFunc(logger, store))
		nodes.GET("/:id/runners", NodeRunnersHandlerFunc(logger, store))
		nodes.POST("", NodeRegisterHandlerFunc(logger, scheduler, store))
		nodes.DELETE("/:id", NodeDeregisterHandlerFunc(logger, scheduler, store))
		nodes.POST("/:id/heartbeat", NodeHeartbeatHandlerFunc(logger, scheduler, store))
		nodes.POST("/:id/cordon", NodeCordonHandlerFunc(logger, scheduler, store))
		nodes.POST("/:id/uncordon", NodeUncordonHandlerFunc(logger, scheduler, store))
	}
}

// NodesHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /api/v1/nodes
func NodesHandlerFunc(logger *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		nodes, err := store.GetNodes(ctx, nil)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(200, gin.H{"nodes": nodes})
	}

	return f
}

// NodeHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /api/v1/nodes/:id
func NodeHandlerFunc(logger *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		node, err := store.GetNode(ctx, ctx.Param("id"))
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(200, node)
	}

	return f
}

// NodeRunnersHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /api/v1/nodes/:id/runners
func NodeRunnersHandlerFunc(logger *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		runners, err := store.GetRunners(ctx, func(r *fireactions.Runner) bool {
			if r.NodeID == nil || r.DeletedAt != nil {
				return false
			}

			return *r.NodeID == ctx.Param("id")
		})
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(200, gin.H{"runners": runners})
	}

	return f
}

// NodeRegisterHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint POST /api/v1/nodes
func NodeRegisterHandlerFunc(logger *zerolog.Logger, scheduler *scheduler.Scheduler, storage store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		var req fireactions.NodeRegisterRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.Error(err)
			return
		}

		node, err := storage.GetNodeByName(ctx, req.Name)
		if err != nil && err != store.ErrNotFound {
			ctx.Error(err)
			return
		}

		if node != nil {
			node.CPU.Capacity = req.CpuCapacity
			node.RAM.Capacity = req.RamCapacity
			node.CPU.OvercommitRatio = req.CpuOvercommitRatio
			node.RAM.OvercommitRatio = req.RamOvercommitRatio
			node.Labels = req.Labels
			node.HeartbeatInterval = req.HeartbeatInterval
			node.UpdatedAt = time.Now()

			err = storage.SaveNode(ctx, node)
			if err != nil {
				ctx.Error(err)
				return
			}

			scheduler.NotifyNodeUpdated(node)
			logger.Info().Msgf("updated node: %s", node.Name)
			ctx.JSON(200, &fireactions.NodeRegistrationInfo{ID: node.ID})
			return
		}

		uuid := uuid.New().String()
		node = &fireactions.Node{
			ID:                uuid,
			Name:              req.Name,
			CPU:               fireactions.NodeResource{Allocated: 0, Capacity: req.CpuCapacity, OvercommitRatio: req.CpuOvercommitRatio},
			RAM:               fireactions.NodeResource{Allocated: 0, Capacity: req.RamCapacity, OvercommitRatio: req.RamOvercommitRatio},
			Status:            fireactions.NodeStatusCordoned,
			Labels:            req.Labels,
			HeartbeatInterval: req.HeartbeatInterval,
			LastHeartbeat:     time.Now(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		err = storage.SaveNode(ctx, node)
		if err != nil {
			ctx.Error(err)
			return
		}

		scheduler.NotifyNodeCreated(node)
		logger.Info().Msgf("registered node: %s", node.Name)
		ctx.JSON(200, &fireactions.NodeRegistrationInfo{ID: node.ID})
	}

	return f
}

// NodeDeregisterHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint DELETE /api/v1/nodes/:id
func NodeDeregisterHandlerFunc(logger *zerolog.Logger, scheduler *scheduler.Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")
		node, err := store.GetNode(ctx, id)
		if err != nil {
			ctx.Error(err)
			return
		}

		err = store.DeleteNode(ctx, node.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		scheduler.NotifyNodeDeleted(node)
		logger.Info().Msgf("deregistered node: %s", node.Name)
		ctx.Status(204)
	}

	return f
}

// NodeCordonHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint POST /api/v1/nodes/:id/cordon
func NodeCordonHandlerFunc(logger *zerolog.Logger, scheduler *scheduler.Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		node, err := store.SetNodeStatus(ctx, ctx.Param("id"), fireactions.NodeStatusCordoned)
		if err != nil {
			ctx.Error(err)
			return
		}

		scheduler.NotifyNodeUpdated(node)
		logger.Info().Msgf("cordoned node: %s", node.Name)
		ctx.Status(204)
	}

	return f
}

// NodeUncordonHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint POST /api/v1/nodes/:id/uncordon
func NodeUncordonHandlerFunc(logger *zerolog.Logger, scheduler *scheduler.Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		node, err := store.SetNodeStatus(ctx, ctx.Param("id"), fireactions.NodeStatusReady)
		if err != nil {
			ctx.Error(err)
			return
		}

		scheduler.NotifyNodeUpdated(node)
		logger.Info().Msgf("uncordoned node: %s", node.Name)
		ctx.Status(204)
	}

	return f
}

// NodeHeartbeatHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint POST /api/v1/nodes/:id/heartbeat
func NodeHeartbeatHandlerFunc(logger *zerolog.Logger, scheduler *scheduler.Scheduler, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		node, err := store.SetNodeLastHeartbeat(ctx, ctx.Param("id"), time.Now())
		if err != nil {
			ctx.Error(err)
			return
		}

		scheduler.NotifyNodeUpdated(node)
		ctx.Status(204)
	}

	return f
}
