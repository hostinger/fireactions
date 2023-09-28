package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/structs"
)

// DisconnectNode updates the Node status to offline, resets resource counters and removes all runners that were
// assigned to it.
func (s *Server) DisconnectNode(ctx context.Context, id string) error {
	n, err := s.Store.GetNode(ctx, id)
	if err != nil {
		return err
	}

	n.Status = structs.NodeStatusOffline
	n.CPU.Allocated = 0
	n.RAM.Allocated = 0
	err = s.Store.UpdateNode(ctx, n)
	if err != nil {
		return err
	}

	// (konradasb): Upon node disconnection, we need to delete all runners that were
	// assigned to it.
	runners, err := s.Store.GetRunners(ctx)
	if err != nil {
		return err
	}

	runners = runners.Filter(func(r *structs.Runner) bool {
		return r.GetNode() == n.Name && (r.Status == structs.RunnerStatusAssigned || r.Status == structs.RunnerStatusAccepted)
	})

	for _, r := range runners {
		err = s.Store.DeleteRunner(ctx, r.ID)
		if err != nil {
			return err
		}

		s.log.Debug().Msgf("removing runner %s due to node %s disconnect event", r.Name, n.Name)
	}

	s.scheduler.UpdateNodeInCache(n)
	s.log.Info().Msgf("updated node %s status", n)

	return nil
}

// ConnectNode updates the Node status to online, resets resource counters and removes all runners that were
// assigned to it.
func (s *Server) ConnectNode(ctx context.Context, id string) error {
	n, err := s.Store.GetNode(ctx, id)
	if err != nil {
		return err
	}

	n.Status = structs.NodeStatusOnline
	n.CPU.Allocated = 0
	n.RAM.Allocated = 0
	err = s.Store.UpdateNode(ctx, n)
	if err != nil {
		return err
	}

	// (konradasb): Upon node reconnection, we need to delete all runners that were
	// assigned to it.
	runners, err := s.Store.GetRunners(ctx)
	if err != nil {
		return err
	}

	runners = runners.Filter(func(r *structs.Runner) bool {
		return r.GetNode() == n.Name && (r.Status == structs.RunnerStatusAssigned || r.Status == structs.RunnerStatusAccepted)
	})

	for _, r := range runners {
		err = s.Store.DeleteRunner(ctx, r.ID)
		if err != nil {
			return err
		}

		s.log.Debug().Msgf("removing runner %s due to node %s reconnect event", r.Name, n.Name)
	}

	s.scheduler.UpdateNodeInCache(n)
	s.log.Info().Msgf("updated node %s status", n)

	return nil
}

// AcceptRunner updates the Runner status to accepted.
func (s *Server) AcceptRunner(ctx context.Context, nodeID, runnerID string) error {
	n, err := s.Store.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}

	r, err := s.Store.GetRunner(ctx, runnerID)
	if err != nil {
		return err
	}

	r.Status = structs.RunnerStatusAccepted
	err = s.Store.UpdateRunner(ctx, r)
	if err != nil {
		return err
	}

	s.log.Info().Msgf("runner %s accepted by node %s", r.Name, n.Name)
	return nil
}

// RejectRunner updates the Runner status to rejected.
func (s *Server) RejectRunner(ctx context.Context, nodeID, runnerID string) error {
	n, err := s.Store.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}

	r, err := s.Store.GetRunner(ctx, runnerID)
	if err != nil {
		return err
	}

	r.Status = structs.RunnerStatusRejected
	err = s.Store.UpdateRunner(ctx, r)
	if err != nil {
		return err
	}

	s.log.Info().Msgf("runner %s rejected by node %s", r.Name, n.Name)
	return nil
}

func (s *Server) CompleteRunner(ctx context.Context, nodeID, runnerID string) error {
	n, err := s.Store.GetNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("error getting node %s: %w", nodeID, err)
	}

	r, err := s.Store.GetRunner(ctx, runnerID)
	if err != nil {
		return fmt.Errorf("error getting runner %s: %w", runnerID, err)
	}

	err = s.Store.DeleteRunner(ctx, r.ID)
	if err != nil {
		return fmt.Errorf("error deleting runner %s: %w", runnerID, err)
	}

	err = s.Store.ReleaseNodeResources(ctx, n.ID, r.Flavor.VCPUs, r.Flavor.GetMemorySizeBytes())
	if err != nil {
		return fmt.Errorf("error releasing node resources: %w", err)
	}

	s.log.Info().Msgf("runner %s completed by node %s", r.Name, n.Name)
	return nil
}

func (s *Server) handleNodeRegister(ctx *gin.Context) {
	var req api.NodeRegisterRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	node, err := s.Store.GetNode(ctx, req.UUID)
	if err != nil && !errors.As(err, &store.ErrNotFound{}) {
		httperr.E(ctx, err)
		return
	}

	nodeExists := node != nil

	n := &structs.Node{
		ID:           req.UUID,
		Name:         req.Name,
		Organisation: req.Organisation,
		Group:        req.Group,
		Status:       structs.NodeStatusUnknown,
		CPU:          structs.Resource{Capacity: int64(req.CpuTotal), Allocated: 0, OvercommitRatio: req.CpuOvercommitRatio},
		RAM:          structs.Resource{Capacity: int64(req.MemTotal), Allocated: 0, OvercommitRatio: req.MemOvercommitRatio},
	}

	switch nodeExists {
	case true:
		n.CreatedAt = node.CreatedAt
		n.UpdatedAt = time.Now()
	case false:
		n.CreatedAt = time.Now()
		n.UpdatedAt = time.Now()
	}

	err = s.Store.UpdateNode(ctx, n)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	if !nodeExists {
		s.log.Info().Msgf("registered new node %s: cpu=%d, mem=%d", n.Name, req.CpuTotal, req.MemTotal)
	}

	s.scheduler.AddNodeToCache(n)
	ctx.JSON(200, gin.H{"message": "Node registered successfully"})
}

func (s *Server) handleNodeDeregister(ctx *gin.Context) {
	id := ctx.Param("id")

	n, err := s.Store.GetNode(ctx, id)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	err = s.Store.DeleteNode(ctx, n.ID)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	s.scheduler.DeleteNodeFromCache(n)

	ctx.JSON(200, gin.H{"message": "Node deregistered successfully"})
	s.log.Info().Msgf("node %s deregistered", n.Name)
}

func (s *Server) handleGetNode(ctx *gin.Context) {
	id := ctx.Param("id")

	node, err := s.Store.GetNode(ctx, id)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, convertNodeToNodeV1(node))
}

func (s *Server) handleGetNodes(ctx *gin.Context) {
	type query struct {
		Organisation string `form:"organisation"`
		Group        string `form:"group"`
	}

	var q query
	err := ctx.BindQuery(&q)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	nodes, err := s.Store.GetNodes(ctx)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	nodes = nodes.Filter(func(n *structs.Node) bool {
		if q.Organisation != "" && n.Organisation != q.Organisation {
			return false
		}

		if q.Group != "" && n.Group != q.Group {
			return false
		}

		return true
	})

	ctx.JSON(200, gin.H{"nodes": convertNodesToNodesV1(nodes...)})
}

func (s *Server) handleNodeConnect(ctx *gin.Context) {
	id := ctx.Param("id")

	err := s.ConnectNode(ctx, id)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "Node connected successfully"})
}

func (s *Server) handleNodeDisconnect(ctx *gin.Context) {
	id := ctx.Param("id")

	err := s.DisconnectNode(ctx, id)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "Node disconnected successfully"})
}

func (s *Server) handleAcceptRunnerAssignment(ctx *gin.Context) {
	err := s.AcceptRunner(ctx, ctx.Param("id"), ctx.Param("runner"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "Runner accepted successfully"})
}

func (s *Server) handleRejectRunnerAssignment(ctx *gin.Context) {
	err := s.RejectRunner(ctx, ctx.Param("id"), ctx.Param("runner"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "Runner rejected successfully"})
}

func (s *Server) handleCompleteRunnerAssignment(ctx *gin.Context) {
	err := s.CompleteRunner(ctx, ctx.Param("id"), ctx.Param("runner"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "Runner completed successfully"})
}

func (s *Server) handleGetNodeAssignments(ctx *gin.Context) {
	n, err := s.Store.GetNode(ctx, ctx.Param("id"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	runners, err := s.Store.GetRunners(ctx)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	runners = runners.Filter(func(r *structs.Runner) bool {
		return r.Status == structs.RunnerStatusAssigned && r.GetNode() == n.Name
	})

	n.UpdatedAt = time.Now()
	err = s.Store.UpdateNode(ctx, n)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	s.scheduler.UpdateNodeInCache(n)

	ctx.JSON(200, gin.H{"runners": convertRunnersToRunnersV1(runners...)})
}

func convertNodeToNodeV1(node *structs.Node) *api.Node {
	n := &api.Node{
		ID:           node.ID,
		Name:         node.Name,
		Organisation: node.Organisation,
		Group:        node.Group,
		Status:       string(node.Status),
		CpuTotal:     node.CPU.Capacity,
		CpuFree:      node.CPU.Available(),
		MemTotal:     node.RAM.Capacity,
		MemFree:      node.RAM.Available(),
		LastSeen:     node.UpdatedAt,
	}

	return n
}

func convertNodesToNodesV1(nodes ...*structs.Node) api.Nodes {
	n := make([]*api.Node, 0, len(nodes))
	for _, node := range nodes {
		n = append(n, convertNodeToNodeV1(node))
	}

	return n
}
