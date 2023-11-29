package reconciler

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/reconciler/cache"
	"github.com/hostinger/fireactions/client/reconciler/locker"
	"github.com/rs/zerolog"
)

type Lister interface {
	List(ctx context.Context) ([]*fireactions.Runner, error)
}

type ListFunc func(ctx context.Context) ([]*fireactions.Runner, error)

func (f ListFunc) List(ctx context.Context) ([]*fireactions.Runner, error) {
	return f(ctx)
}

type Syncer interface {
	Sync(ctx context.Context, runner *fireactions.Runner) error
}

type Reconciler struct {
	lister        Lister
	locker        locker.Locker
	interval      time.Duration
	syncer        Syncer
	maxConcurrent int
	queue         chan string
	cache         cache.Cache
	logger        *zerolog.Logger
}

// Opt is a functional option for the Reconciler.
type Opt func(*Reconciler)

// WithInterval sets the reconcile interval for the Reconciler.
func WithInterval(interval time.Duration) Opt {
	f := func(r *Reconciler) {
		r.interval = interval
	}

	return f
}

// WithMaxConcurrent sets the maximum number of concurrent workers for the Reconciler.
func WithMaxConcurrent(maxConcurrent int) Opt {
	f := func(r *Reconciler) {
		r.maxConcurrent = maxConcurrent
	}

	return f
}

// WithLogger sets the logger for the Reconciler.
func WithLogger(logger *zerolog.Logger) Opt {
	f := func(r *Reconciler) {
		r.logger = logger
	}

	return f
}

// WithLocker sets the locker for the Reconciler.
func WithLocker(locker locker.Locker) Opt {
	f := func(r *Reconciler) {
		r.locker = locker
	}

	return f
}

// WithCache sets the cache for the Reconciler.
func WithCache(cache cache.Cache) Opt {
	f := func(r *Reconciler) {
		r.cache = cache
	}

	return f
}

// NewReconciler creates a new Reconciler.
func NewReconciler(lister Lister, syncer Syncer, opts ...Opt) *Reconciler {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	r := &Reconciler{
		lister:        lister,
		syncer:        syncer,
		locker:        locker.NewMemoryLocker(),
		interval:      10 * time.Second,
		maxConcurrent: 10,
		cache:         cache.NewCache(),
		queue:         make(chan string, 500),
		logger:        &logger,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Run starts the Reconciler.
func (r *Reconciler) Run(ctx context.Context) {
	r.logger.Info().Msgf("reconciler: starting reconciliation loop")
	go r.runReconciliationLoop(ctx)

	for i := 0; i < r.maxConcurrent; i++ {
		r.logger.Info().Msgf("reconciler: starting worker %d", i)
		go r.runWorker(ctx)
	}

	<-ctx.Done()
}

func (r *Reconciler) runReconciliationLoop(ctx context.Context) {
	for {
		r.reconcile(ctx)

		select {
		case <-time.After(r.interval):
		case <-ctx.Done():
			return
		}
	}
}

func (r *Reconciler) reconcile(ctx context.Context) {
	runners, err := r.lister.List(ctx)
	if err != nil {
		r.logger.Err(err).Msgf("reconciler: failed to execute lister")
		return
	}

	queued := 0
	for _, runner := range runners {
		cached, ok := r.cache.Get(runner.ID)
		if ok && cached.GetRunner().Equals(*runner) {
			continue
		}

		r.cache.Set(runner.ID, runner)
		r.queue <- runner.ID
		queued++
	}

	r.logger.Debug().Msgf("reconciler: queued %d runners", queued)
}

func (r *Reconciler) runWorker(ctx context.Context) {
	for {
		var runnerID string
		select {
		case <-ctx.Done():
			return
		case runnerID = <-r.queue:
		}

		r.syncRunner(ctx, runnerID)
	}
}

func (r *Reconciler) syncRunner(ctx context.Context, runnerID string) {
	if err := r.locker.Acquire(ctx, runnerID); err != nil {
		return
	}
	defer r.locker.Release(ctx, runnerID)

	runner, ok := r.cache.Get(runnerID)
	if !ok {
		return
	}

	err := r.syncer.Sync(ctx, runner.GetRunner())
	switch errors.Unwrap(err) {
	case context.Canceled:
		return
	case nil:
		r.cache.Delete(runnerID)
		r.logger.Debug().Msgf("reconciler: runner %s synced", runnerID)
		return
	}

	r.cache.IncAttempts(runnerID)
	attempts := r.cache.GetAttempts(runnerID)

	go func() {
		select {
		case <-time.After(time.Duration(attempts) * time.Second):
		case <-ctx.Done():
			return
		}

		r.queue <- runnerID
	}()

	r.logger.Err(err).Msgf("reconciler: failed to sync runner %s, requeueing in %d seconds (attempt %d)", runnerID, attempts, r.cache.GetAttempts(runnerID))
}
