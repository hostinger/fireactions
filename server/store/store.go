//go:generate mockgen -source=store.go -destination=mock/store.go -package=mock
package store

import (
	"context"

	"github.com/google/go-github/v50/github"
	"github.com/hostinger/fireactions"
)

// Store is a common interface for all storage backend implementations.
type Store interface {
	GetNodes(ctx context.Context, filter fireactions.NodeFilterFunc) ([]*fireactions.Node, error)
	GetNode(ctx context.Context, id string) (*fireactions.Node, error)
	SaveNode(ctx context.Context, node *fireactions.Node) error
	GetNodeByName(ctx context.Context, name string) (*fireactions.Node, error)
	DeleteNode(ctx context.Context, id string) error
	UpdateNodeWithTransaction(ctx context.Context, tx Tx, id string, updateFunc func(*fireactions.Node) error) (*fireactions.Node, error)
	UpdateNode(ctx context.Context, id string, updateFunc func(*fireactions.Node) error) (*fireactions.Node, error)

	GetRunners(ctx context.Context, filter fireactions.RunnerFilterFunc) ([]*fireactions.Runner, error)
	GetRunner(ctx context.Context, id string) (*fireactions.Runner, error)
	SaveRunner(ctx context.Context, runner *fireactions.Runner) error
	UpdateRunnerWithTransaction(ctx context.Context, tx Tx, id string, runnerUpdateFn func(*fireactions.Runner) error) (*fireactions.Runner, error)
	UpdateRunner(ctx context.Context, id string, runnerUpdateFn func(*fireactions.Runner) error) (*fireactions.Runner, error)
	DeleteRunner(ctx context.Context, id string) error

	SaveWorkflowRun(ctx context.Context, workflowRun *github.WorkflowRun) error
	GetWorkflowRuns(ctx context.Context, filter func(*github.WorkflowRun) bool) ([]*github.WorkflowRun, error)
	GetWorkflowRun(ctx context.Context, id int64) (*github.WorkflowRun, error)
	DeleteWorkflowRun(ctx context.Context, id int64) error

	GetWorkflowJob(ctx context.Context, runID int64, id int64) (*github.WorkflowJob, error)
	SaveWorkflowJob(ctx context.Context, workflowJob *github.WorkflowJob) error
	GetWorkflowJobs(ctx context.Context, runID int64, filter func(*github.WorkflowJob) bool) ([]*github.WorkflowJob, error)
	DeleteWorkflowJob(ctx context.Context, runID int64, id int64) error

	BeginTransaction() (Tx, error)
	Close() error
}

type Tx interface {
	Commit() error
	Rollback() error
}
