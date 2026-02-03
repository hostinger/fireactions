package agent

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hostinger/fireactions/agent/tail"
	agentv1 "github.com/hostinger/fireactions/proto/agent/v1"
)

func (a *Agent) GetRunnerState(ctx context.Context, req *agentv1.GetRunnerStateRequest) (*agentv1.GetRunnerStateResponse, error) {
	if a.runner == nil {
		return nil, fmt.Errorf("runner not initialized")
	}

	resp := &agentv1.GetRunnerStateResponse{
		State: string(a.runner.GetState()),
	}

	return resp, nil
}

func (a *Agent) GetRunnerVersion(ctx context.Context, req *agentv1.GetRunnerVersionRequest) (*agentv1.GetRunnerVersionResponse, error) {
	version, err := a.runner.GetVersion()
	if err != nil {
		return nil, fmt.Errorf("get runner version: %w", err)
	}

	resp := &agentv1.GetRunnerVersionResponse{
		Version: version,
	}

	return resp, nil
}

func (a *Agent) GetLogs(req *agentv1.GetLogsRequest, stream agentv1.AgentService_GetLogsServer) error {
	ctx := stream.Context()

	if _, err := os.Stat(a.logFile); os.IsNotExist(err) && !req.Follow {
		return nil
	}

	config := tail.Config{
		Follow: req.Follow,
		ReOpen: req.Follow,
	}

	if req.TailLines > 0 {
		config.Location = &tail.SeekInfo{Offset: int64(-req.TailLines), Whence: io.SeekEnd}
	}

	t, err := tail.TailFile(a.logFile, config)
	if err != nil {
		return fmt.Errorf("tail file: %w", err)
	}
	defer t.Cleanup()

	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return nil
		case line, ok := <-t.Lines:
			if !ok {
				if t.Err() != nil {
					return fmt.Errorf("tail error: %w", t.Err())
				}
				return nil
			}
			if line.Err != nil {
				a.logger.Warn().Err(line.Err).Msg("Error reading log line")
				continue
			}
			if err := stream.Send(&agentv1.GetLogsResponse{
				Line: line.Text + "\n",
			}); err != nil {
				t.Stop()
				return err
			}
		}
	}
}
