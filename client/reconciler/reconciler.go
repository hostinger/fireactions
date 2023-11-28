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
	queue         chan *fireactions.Runner
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
		queue:         make(chan *fireactions.Runner, 500),
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
		_, ok := r.cache.Get(runner.ID)
		if ok {
			r.logger.Debug().Msgf("reconciler: runner %s already queued, skipping", runner.ID)
			continue
		}

		r.cache.Set(runner.ID, runner)

		ok = r.tryEnqueue(runner)
		if !ok {
			r.logger.Debug().Msgf("reconciler: queue full, skipping runner %s", runner.ID)
			continue
		}

		queued++
	}

	r.logger.Debug().Msgf("reconciler: queued %d runners", queued)
}

func (r *Reconciler) runWorker(ctx context.Context) {
	for {
		var runner *fireactions.Runner
		select {
		case <-ctx.Done():
			return
		case runner = <-r.queue:
		}

		r.syncRunner(ctx, runner)
	}
}

func (r *Reconciler) syncRunner(ctx context.Context, runner *fireactions.Runner) {
	if err := r.locker.Acquire(ctx, runner); err != nil {
		return
	}
	defer r.locker.Release(ctx, runner)

	err := r.syncer.Sync(ctx, runner)
	switch errors.Unwrap(err) {
	case context.Canceled:
		return
	case nil:
		r.cache.Delete(runner.ID)
		r.logger.Debug().Msgf("reconciler: runner %s synced", runner.ID)
		return
	}

	r.cache.IncAttempts(runner.ID)
	attempts := r.cache.GetAttempts(runner.ID)

	go func() {
		select {
		case <-time.After(time.Duration(attempts) * time.Second):
			r.queue <- runner
		case <-ctx.Done():
			return
		}

		r.queue <- runner
	}()

	r.logger.Err(err).Msgf("reconciler: failed to sync runner %s, requeueing in %d seconds (attempt %d)", runner.ID, attempts, r.cache.GetAttempts(runner.ID))
}

func (r *Reconciler) tryEnqueue(runner *fireactions.Runner) bool {
	select {
	case r.queue <- runner:
		return true
	default:
		return false
	}
}
