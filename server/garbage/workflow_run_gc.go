package garbage

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/rs/zerolog"
)

var (
	defaultWorkflowRunGCInterval = 1 * time.Hour
	defaultMaxWorkflowRunAge     = 24 * time.Hour * 31
)

// WorkflowJobLister is an interface for listing workflow jobs.
type WorkflowRunLister interface {
	GetWorkflowRuns(ctx context.Context, filter func(*github.WorkflowRun) bool) ([]*github.WorkflowRun, error)
}

// WorkflowJobDeleter is an interface for deleting workflow jobs.
type WorkflowRunDeleter interface {
	DeleteWorkflowRun(ctx context.Context, id int64) error
}

type workflowRunGC struct {
	interval          time.Duration
	maxWorkflowRunAge time.Duration
	lister            WorkflowRunLister
	deleter           WorkflowRunDeleter
	logger            *zerolog.Logger
	lock              sync.Mutex
}

// WorkflowRunGCOption is an option for configuring workflow run GC.
type WorkflowRunGCOption func(*workflowRunGC)

// WithWorkflowRunGCInterval configures the interval at which workflow run GC runs.
func WithWorkflowRunGCInterval(interval time.Duration) WorkflowRunGCOption {
	f := func(g *workflowRunGC) {
		g.interval = interval
	}

	return f
}

// WithMaxWorkflowRunAge configures the maximum age of workflow runs to be garbage collected.
func WithMaxWorkflowRunAge(maxWorkflowRunAge time.Duration) WorkflowRunGCOption {
	f := func(g *workflowRunGC) {
		g.maxWorkflowRunAge = maxWorkflowRunAge
	}

	return f
}

// WithWorkflowRunGCLogger configures the logger for workflow run GC.
func WithWorkflowRunGCLogger(logger *zerolog.Logger) WorkflowRunGCOption {
	f := func(g *workflowRunGC) {
		g.logger = logger
	}

	return f
}

// NewWorkflowRunGC creates a new workflow run GC.
func NewWorkflowRunGC(lister WorkflowRunLister, deleter WorkflowRunDeleter, opts ...WorkflowRunGCOption) *workflowRunGC {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	gc := &workflowRunGC{
		interval:          defaultWorkflowRunGCInterval,
		maxWorkflowRunAge: defaultMaxWorkflowRunAge,
		lister:            lister,
		deleter:           deleter,
		logger:            &logger,
		lock:              sync.Mutex{},
	}

	for _, opt := range opts {
		opt(gc)
	}

	return gc
}

// Run starts the workflow run GC. It blocks until the context is canceled.
func (g *workflowRunGC) Run(ctx context.Context) {
	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		err := g.GarbageCollect(ctx)
		if err != nil {
			g.logger.Error().Err(err).Msg("gc: workflow_run: failed to garbage collect")
		}
	}
}

// GarbageCollect runs the workflow run GC once.
func (g *workflowRunGC) GarbageCollect(ctx context.Context) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.logger.Info().Msg("gc: workflow_run: starting garbage collection")
	err := g.gc(ctx)
	if err != nil {
		return err
	}

	g.logger.Info().Msg("gc: workflow_run: finished garbage collection")
	return nil
}

func (g *workflowRunGC) gc(ctx context.Context) error {
	workflowRuns, err := g.lister.GetWorkflowRuns(ctx, func(wr *github.WorkflowRun) bool {
		return wr.GetUpdatedAt().Before(time.Now().Add(-g.maxWorkflowRunAge))
	})
	if err != nil {
		return err
	}

	if len(workflowRuns) == 0 {
		return nil
	}

	for _, wr := range workflowRuns {
		if err := g.deleter.DeleteWorkflowRun(ctx, wr.GetID()); err != nil {
			return err
		}

		g.logger.Info().Int64("id", wr.GetID()).Msgf("gc: workflow_run: deleted workflow run %d", wr.GetID())
	}

	return nil
}
