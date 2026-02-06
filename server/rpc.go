package server

import (
	"context"
	"io"
	"sort"

	agentv1 "github.com/hostinger/fireactions/proto/agent/v1"
	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetPool implements ServerService.GetPool.
func (s *Server) GetPool(ctx context.Context, req *serverv1.GetPoolRequest) (*serverv1.GetPoolResponse, error) {
	pool, err := s.findPool(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "pool not found: %v", err)
	}

	return &serverv1.GetPoolResponse{Pool: convertPoolToProto(ctx, pool)}, nil
}

// ListPools implements ServerService.ListPools.
func (s *Server) ListPools(ctx context.Context, req *serverv1.ListPoolsRequest) (*serverv1.ListPoolsResponse, error) {
	s.l.Lock()
	pools := make([]*Pool, 0, len(s.pools))
	for _, pool := range s.pools {
		pools = append(pools, pool)
	}
	s.l.Unlock()

	// Sort pools by name
	sort.Slice(pools, func(i, j int) bool {
		return pools[i].config.Name < pools[j].config.Name
	})

	// Convert to proto messages
	protoPools := make([]*serverv1.Pool, len(pools))
	for i, pool := range pools {
		protoPools[i] = convertPoolToProto(ctx, pool)
	}

	return &serverv1.ListPoolsResponse{Pools: protoPools}, nil
}

// ScalePool implements ServerService.ScalePool.
func (s *Server) ScalePool(ctx context.Context, req *serverv1.ScalePoolRequest) (*serverv1.ScalePoolResponse, error) {
	pool, err := s.findPool(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "pool not found: %v", err)
	}

	metricPoolScaleRequests.WithLabelValues(req.Name, pool.config.Runner.Organization).Inc()

	// Update the pool config with the new replicas value
	// The Run() loop will handle the actual scaling
	pool.SetReplicas(int(req.Replicas))

	return &serverv1.ScalePoolResponse{Message: "Pool replicas updated successfully"}, nil
}

// PausePool implements ServerService.PausePool.
func (s *Server) PausePool(ctx context.Context, req *serverv1.PausePoolRequest) (*serverv1.PausePoolResponse, error) {
	pool, err := s.findPool(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "pool not found: %v", err)
	}

	pool.Pause()
	metricPoolStatus.WithLabelValues(req.Name).Set(0)

	return &serverv1.PausePoolResponse{Message: "Pool paused successfully"}, nil
}

// ResumePool implements ServerService.ResumePool.
func (s *Server) ResumePool(ctx context.Context, req *serverv1.ResumePoolRequest) (*serverv1.ResumePoolResponse, error) {
	pool, err := s.findPool(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "pool not found: %v", err)
	}

	pool.Resume()
	metricPoolStatus.WithLabelValues(req.Name).Set(1)

	return &serverv1.ResumePoolResponse{Message: "Pool resumed successfully"}, nil
}

// ListMachines implements ServerService.ListMachines.
func (s *Server) ListMachines(ctx context.Context, req *serverv1.ListMachinesRequest) (*serverv1.ListMachinesResponse, error) {
	var machines []*Machine

	if req.Pool == "" {
		// List all machines across all pools
		s.l.Lock()
		pools := make([]*Pool, 0, len(s.pools))
		for _, pool := range s.pools {
			pools = append(pools, pool)
		}
		s.l.Unlock()

		for _, pool := range pools {
			poolMachines, err := pool.ListMachines(ctx)
			if err != nil {
				s.logger.Warn().Err(err).Str("pool", pool.config.Name).Msg("Failed to list machines for pool")
				continue
			}

			machines = append(machines, poolMachines...)
		}
	} else {
		// List machines for specific pool
		pool, err := s.findPool(req.Pool)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "pool not found: %v", err)
		}

		poolMachines, err := pool.ListMachines(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "list machines: %v", err)
		}

		machines = poolMachines
	}

	if machines == nil {
		machines = []*Machine{}
	}

	sort.Slice(machines, func(i, j int) bool {
		if machines[i].Pool != machines[j].Pool {
			return machines[i].Pool < machines[j].Pool
		}

		return machines[i].Name < machines[j].Name
	})

	// Convert machines to proto in parallel for better performance
	protoMachines := make([]*serverv1.Machine, len(machines))
	type result struct {
		index int
		proto *serverv1.Machine
	}
	results := make(chan result, len(machines))

	for i, machine := range machines {
		go func(idx int, m *Machine) {
			results <- result{index: idx, proto: convertMachineToProto(ctx, m)}
		}(i, machine)
	}

	// Collect results
	for range machines {
		r := <-results
		protoMachines[r.index] = r.proto
	}

	return &serverv1.ListMachinesResponse{Machines: protoMachines}, nil
}

// GetMachine implements ServerService.GetMachine.
func (s *Server) GetMachine(ctx context.Context, req *serverv1.GetMachineRequest) (*serverv1.GetMachineResponse, error) {
	machine, err := s.findMachine(req.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "machine not found: %v", err)
	}

	return &serverv1.GetMachineResponse{Machine: convertMachineToProto(ctx, machine)}, nil
}

// GetHealth implements ServerService.GetHealth.
func (s *Server) GetHealth(ctx context.Context, req *serverv1.GetHealthRequest) (*serverv1.GetHealthResponse, error) {
	return &serverv1.GetHealthResponse{Status: "OK"}, nil
}

// GetVersion implements ServerService.GetVersion.
func (s *Server) GetVersion(ctx context.Context, req *serverv1.GetVersionRequest) (*serverv1.GetVersionResponse, error) {
	return &serverv1.GetVersionResponse{Version: s.version, Commit: s.commit, Date: s.date}, nil
}

// GetMachineLogs implements ServerService.GetMachineLogs (server-side streaming).
func (s *Server) GetMachineLogs(req *serverv1.GetMachineLogsRequest, stream serverv1.ServerService_GetMachineLogsServer) error {
	ctx := stream.Context()

	machine, err := s.findMachine(req.GetID())
	if err != nil {
		return status.Errorf(codes.NotFound, "machine not found: %v", err)
	}

	conn, client, err := machine.ConnectToGuestAgent(ctx)
	if err != nil {
		return status.Errorf(codes.Internal, "connect to agent: %v", err)
	}
	defer conn.Close()

	agentStream, err := client.GetLogs(ctx, &agentv1.GetLogsRequest{
		Follow:    req.Follow,
		TailLines: req.TailLines,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "agent GetLogs: %v", err)
	}

	// Proxy logs from agent to CLI
	for {
		agentResp, err := agentStream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "receive from agent: %v", err)
		}

		// Forward to client
		if err := stream.Send(&serverv1.GetMachineLogsResponse{Line: agentResp.Line}); err != nil {
			return err
		}
	}
}

// ListImages implements ServerService.ListImages.
func (s *Server) ListImages(ctx context.Context, req *serverv1.ListImagesRequest) (*serverv1.ListImagesResponse, error) {
	images, err := s.imageManager.listImages(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list images: %v", err)
	}

	protoImages := make([]*serverv1.Image, len(images))
	for i, img := range images {
		protoImages[i] = convertImageToProto(ctx, img)
	}

	return &serverv1.ListImagesResponse{Images: protoImages}, nil
}

// RemoveImage implements ServerService.RemoveImage.
func (s *Server) RemoveImage(ctx context.Context, req *serverv1.RemoveImageRequest) (*serverv1.RemoveImageResponse, error) {
	if err := s.imageManager.removeImage(ctx, req.Name); err != nil {
		return nil, status.Errorf(codes.Internal, "remove image: %v", err)
	}

	return &serverv1.RemoveImageResponse{Message: "Image removed successfully"}, nil
}
