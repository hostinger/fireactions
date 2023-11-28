package server

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/helper/stringid"
	"github.com/hostinger/fireactions/server/store"
)

func (s *Server) handleRegisterNode(ctx *gin.Context) {
	var req fireactions.NodeRegisterRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Bad request: %s", err.Error())})
		return
	}

	node, err := s.store.GetNodeByName(ctx, req.Name)
	if err != nil && err != store.ErrNotFound {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	if node != nil {
		node.CPU.Capacity = req.CpuCapacity
		node.RAM.Capacity = req.RamCapacity
		node.CPU.OvercommitRatio = req.CpuOvercommitRatio
		node.RAM.OvercommitRatio = req.RamOvercommitRatio
		node.Labels = req.Labels
		node.ReconcileInterval = req.ReconcileInterval
		node.UpdatedAt = time.Now()

		err = s.store.SaveNode(ctx, node)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
			return
		}

		runners, err := s.store.GetRunners(ctx, func(r *fireactions.Runner) bool {
			return r.GetNodeID() == node.ID && r.Status.State != fireactions.RunnerStateCompleted
		})
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
			return
		}

		for _, runner := range runners {
			_, err = s.store.UpdateRunner(ctx, runner.ID, func(r *fireactions.Runner) error {
				r.Status = fireactions.RunnerStatus{State: fireactions.RunnerStateCompleted, Description: "Completed due to node re-registration"}
				return nil
			})
			if err != nil {
				ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
				return
			}

			s.logger.Warn().Msgf("completed runner %s due to node re-registration", runner.ID)
		}

		s.scheduler.NotifyNodeUpdated(node)
		s.logger.Info().Msgf("re-registered node: %s", node.Name)
		ctx.JSON(200, &fireactions.NodeRegistrationInfo{ID: node.ID})
		return
	}

	node = &fireactions.Node{
		ID:                stringid.New(),
		Name:              req.Name,
		CPU:               fireactions.NodeResource{Allocated: 0, Capacity: req.CpuCapacity, OvercommitRatio: req.CpuOvercommitRatio},
		RAM:               fireactions.NodeResource{Allocated: 0, Capacity: req.RamCapacity, OvercommitRatio: req.RamOvercommitRatio},
		Status:            fireactions.NodeStatusCordoned,
		Labels:            req.Labels,
		ReconcileInterval: req.ReconcileInterval,
		LastReconcileAt:   time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	err = s.store.SaveNode(ctx, node)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	s.scheduler.NotifyNodeCreated(node)
	s.logger.Info().Msgf("registered node: %s", node.Name)

	ctx.JSON(200, &fireactions.NodeRegistrationInfo{ID: node.ID})
}

func (s *Server) handleCordonNode(ctx *gin.Context) {
	node, err := s.store.UpdateNode(ctx, ctx.Param("id"), func(n *fireactions.Node) error {
		n.Status = fireactions.NodeStatusCordoned
		return nil
	})
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	s.scheduler.NotifyNodeUpdated(node)
	s.logger.Info().Msgf("cordoned node: %s", node.Name)

	ctx.Status(204)
}

func (s *Server) handleUncordonNode(ctx *gin.Context) {
	node, err := s.store.UpdateNode(ctx, ctx.Param("id"), func(n *fireactions.Node) error {
		n.Status = fireactions.NodeStatusReady
		return nil
	})
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	s.scheduler.NotifyNodeUpdated(node)
	s.logger.Info().Msgf("uncordoned node: %s", node.Name)

	ctx.Status(204)
}

func (s *Server) handleDeregisterNode(ctx *gin.Context) {
	id := ctx.Param("id")
	node, err := s.store.GetNode(ctx, id)
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", id)})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	err = s.store.DeleteNode(ctx, node.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	s.scheduler.NotifyNodeDeleted(node)
	s.logger.Info().Msgf("deregistered node: %s", node.Name)
	ctx.Status(204)
}

func (s *Server) handleGetNodeRunners(ctx *gin.Context) {
	node, err := s.store.UpdateNode(ctx, ctx.Param("id"), func(n *fireactions.Node) error {
		n.LastReconcileAt = time.Now()
		return nil
	})
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	runners, err := s.store.GetRunners(ctx, func(r *fireactions.Runner) bool {
		if r.NodeID == nil || r.DeletedAt != nil {
			return false
		}

		return *r.NodeID == node.ID
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	s.scheduler.NotifyNodeUpdated(node)
	ctx.JSON(200, gin.H{"runners": runners})
}

func (s *Server) handleGetNode(ctx *gin.Context) {
	node, err := s.store.GetNode(ctx, ctx.Param("id"))
	if err != nil {
		if err == store.ErrNotFound {
			ctx.AbortWithStatusJSON(404, gin.H{"error": fmt.Sprintf("Node with ID %s doesn't exist", ctx.Param("id"))})
		} else {
			ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		}

		return
	}

	ctx.JSON(200, node)
}

func (s *Server) handleGetNodes(ctx *gin.Context) {
	nodes, err := s.store.GetNodes(ctx, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("Internal server error: %s", err.Error())})
		return
	}

	ctx.JSON(200, gin.H{"nodes": nodes})
}
