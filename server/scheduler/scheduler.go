package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/scheduler/cache"
	"github.com/hostinger/fireactions/server/scheduler/filter"
	"github.com/hostinger/fireactions/server/scheduler/filter/cordon"
	"github.com/hostinger/fireactions/server/scheduler/filter/cpucapacity"
	"github.com/hostinger/fireactions/server/scheduler/filter/group"
	"github.com/hostinger/fireactions/server/scheduler/filter/heartbeat"
	"github.com/hostinger/fireactions/server/scheduler/filter/organisation"
	"github.com/hostinger/fireactions/server/scheduler/filter/ramcapacity"
	"github.com/hostinger/fireactions/server/scheduler/filter/status"
	"github.com/hostinger/fireactions/server/scheduler/scorer"
	"github.com/hostinger/fireactions/server/scheduler/scorer/freecpu"
	"github.com/hostinger/fireactions/server/scheduler/scorer/freeram"
	"github.com/rs/zerolog"
)

// Storer is an interface that stores and retrieves Runners and Nodes.
type Storer interface {
	ListRunners(ctx context.Context) ([]*models.Runner, error)
	SaveRunner(ctx context.Context, runner *models.Runner) error
	ListNodes(ctx context.Context) ([]*models.Node, error)
	ReserveNodeResources(ctx context.Context, nodeID string, vcpus int64, ram int64) error
}

// Scheduler is responsible for scheduling Runners onto Nodes.
type Scheduler struct {
	queue        *SchedulingQueue
	cache        cache.Cache
	store        Storer
	filters      map[string]filter.Filter
	scorers      map[string]scorer.Scorer
	shutdownOnce sync.Once
	shutdownCh   chan struct{}
	isShutdown   bool
	config       *Config
	logger       *zerolog.Logger
}

// New creates a new Scheduler.
func New(logger zerolog.Logger, cfg *Config, store Storer) (*Scheduler, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	logger = logger.With().Str("subsystem", "scheduler").Logger()
	s := &Scheduler{
		queue:        NewSchedulingQueue(),
		filters:      make(map[string]filter.Filter, 0),
		scorers:      make(map[string]scorer.Scorer, 0),
		cache:        cache.New(),
		store:        store,
		shutdownOnce: sync.Once{},
		shutdownCh:   make(chan struct{}),
		isShutdown:   false,
		config:       cfg,
		logger:       &logger,
	}

	s.registerFilters()
	s.registerScorers()

	return s, nil
}

// Start starts the Scheduler and runs until Shutdown() is called.
func (s *Scheduler) Start() error {
	err := s.init()
	if err != nil {
		return fmt.Errorf("error initializing scheduler: %w", err)
	}

	go s.runScheduleLoop()
	return nil
}

// Shutdown shuts down the Scheduler.
func (s *Scheduler) Shutdown() {
	s.shutdownOnce.Do(func() { s.isShutdown = true; close(s.shutdownCh) })
}

// Schedule places a Runner into the scheduling queue. If the Runner is already
// in the queue, an error is returned.
func (s *Scheduler) Schedule(r *models.Runner) error {
	if s.isShutdown {
		return nil
	}

	err := s.queue.Enqueue(r)
	if err != nil {
		return fmt.Errorf("error enqueueing runner: %w", err)
	}

	return nil
}

func (s *Scheduler) registerFilters() {
	filters := []filter.Filter{
		cordon.New(), organisation.New(),
		cpucapacity.New(), ramcapacity.New(), group.New(), heartbeat.New(), status.New(),
	}

	for _, filter := range filters {
		_, ok := s.filters[filter.Name()]
		if ok {
			panic(fmt.Sprintf("filter %s already exists", filter.Name()))
		}

		s.filters[filter.Name()] = filter
		s.logger.Debug().Msgf("registered filter %s", filter)
	}
}

func (s *Scheduler) registerScorers() {
	scorers := []scorer.Scorer{
		freecpu.New(s.config.FreeCpuScorerMultiplier), freeram.New(s.config.FreeRamScorerMultiplier),
	}

	for _, scorer := range scorers {
		_, ok := s.scorers[scorer.Name()]
		if ok {
			panic(fmt.Sprintf("scorer %s already exists", scorer.Name()))
		}

		s.scorers[scorer.Name()] = scorer
		s.logger.Debug().Msgf("registered scorer %s", scorer)
	}
}

func (s *Scheduler) init() error {
	nodes, err := s.store.ListNodes(context.Background())
	if err != nil {
		return err
	}

	for _, n := range nodes {
		err = s.cache.AddNode(n)
		if err != nil {
			return err
		}

		s.logger.Debug().Msgf("added existing node %s to internal cache", n)
	}

	runners, err := s.store.ListRunners(context.Background())
	if err != nil {
		return err
	}

	runners = models.FilterRunners(runners, func(r *models.Runner) bool {
		return r.Status == models.RunnerStatusPending
	})

	for _, r := range runners {
		err := s.queue.Enqueue(r)
		if err != nil {
			return err
		}

		s.logger.Debug().Msgf("added existing runner %s to scheduling queue", r)
	}

	return nil
}

func (s *Scheduler) runScheduleLoop() {
	for {
		select {
		case <-s.shutdownCh:
			return
		default:
			s.schedule()
		}
	}
}

func (s *Scheduler) schedule() {
	runner, err := s.queue.Dequeue()
	switch err {
	case nil:
	case ErrQueueClosed:
		return
	default:
		s.logger.Error().Err(err).Msg("error dequeuing runner")
		return
	}

	cache := s.cache.DeepCopy().(cache.Cache)

	nodes, err := cache.GetNodes()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting nodes from cache")
		return
	}

	feasibleNodes := findFeasibleNodes(runner, nodes, s.filters)
	if len(feasibleNodes) == 0 {
		s.queue.Block(runner.ID)
		return
	}

	bestNode := findBestNode(runner, feasibleNodes, s.scorers)
	if bestNode == nil {
		s.queue.Block(runner.ID)
		return
	}

	runner.Status = models.RunnerStatusAssigned
	runner.Node = bestNode
	err = s.store.SaveRunner(context.Background(), runner)
	if err != nil {
		s.logger.Error().Err(err).Msg("error updating runner")
		return
	}

	err = s.store.ReserveNodeResources(context.Background(), bestNode.ID, runner.Flavor.VCPUs, runner.Flavor.GetMemorySizeBytes())
	if err != nil {
		s.logger.Error().Err(err).Msg("error reserving node resources")
		return
	}

	s.logger.Info().Msgf("runner %s is assigned to node %s", runner.ID, bestNode.ID)
}

func findFeasibleNodes(runner *models.Runner, nodes []*models.Node, filters map[string]filter.Filter) []*models.Node {
	feasible := make([]*models.Node, 0, len(nodes))
	for _, n := range nodes {
		if !runFilters(runner, n, filters) {
			continue
		}
		feasible = append(feasible, n)
	}

	return feasible
}

func findBestNode(runner *models.Runner, nodes []*models.Node, scorers map[string]scorer.Scorer) *models.Node {
	if len(nodes) == 0 {
		return nil
	}

	nodesMap := make(map[string]*models.Node, len(nodes))
	for _, n := range nodes {
		nodesMap[n.ID] = n
	}

	scores := make(map[string]float64, len(nodes))
	for _, n := range nodes {
		result, err := runScorers(runner, n, scorers)
		if err != nil {
			continue
		}

		scores[n.ID] = result
	}

	var bestNode *models.Node
	var bestScore float64

	for nodeID, score := range scores {
		if score < bestScore {
			continue
		}

		bestNode, bestScore = nodesMap[nodeID], score
	}

	return bestNode
}

func runScorers(runner *models.Runner, node *models.Node, scorers map[string]scorer.Scorer) (float64, error) {
	var score float64
	for _, scorer := range scorers {
		result, err := scorer.Score(runner, node)
		if err != nil {
			return 0, err
		}

		score += result
	}

	return score, nil
}

func runFilters(runner *models.Runner, node *models.Node, filters map[string]filter.Filter) bool {
	for _, filter := range filters {
		ok, err := filter.Filter(context.Background(), runner, node)
		if err != nil {
			return false
		}

		if ok {
			continue
		}

		return false
	}

	return true
}
