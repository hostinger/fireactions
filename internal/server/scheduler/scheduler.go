package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/hostinger/fireactions/internal/server/scheduler/cache"
	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
)

var (
	// ErrFilterExists is returned when a Filter with the same name already
	// exists.
	ErrFilterExists = fmt.Errorf("filter already exists")
	// ErrScorerExists is returned when a Scorer with the same name already
	// exists.
	ErrScorerExists = fmt.Errorf("scorer already exists")
)

// Scheduler is responsible for scheduling Runners onto Nodes.
type Scheduler struct {
	queue        *SchedulingQueue
	cache        Cache
	store        store.Store
	filters      map[string]Filter
	scorers      map[string]Scorer
	shutdownOnce sync.Once
	shutdownCh   chan struct{}
	isShutdown   bool
	log          *zerolog.Logger
}

// New creates a new Scheduler.
func New(log *zerolog.Logger, cfg *Config, store store.Store) *Scheduler {
	if cfg.FreeCpuScorerMultiplier == 0 {
		cfg.FreeCpuScorerMultiplier = defaultCpuScorerMultiplier
	}

	if cfg.FreeMemScorerMultiplier == 0 {
		cfg.FreeMemScorerMultiplier = defaultRamScorerMultiplier
	}

	s := &Scheduler{
		queue:        NewSchedulingQueue(),
		filters:      make(map[string]Filter, 0),
		scorers:      make(map[string]Scorer, 0),
		cache:        cache.New(),
		store:        store,
		shutdownOnce: sync.Once{},
		shutdownCh:   make(chan struct{}),
		isShutdown:   false,
		log:          log,
	}

	logger := log.With().Str("subsystem", "scheduler").Logger()
	s.log = &logger

	s.MustRegisterScorer(&FreeCpuScorer{
		Multiplier: cfg.FreeCpuScorerMultiplier})
	s.MustRegisterScorer(&FreeRamScorer{
		Multiplier: cfg.FreeMemScorerMultiplier})

	s.MustRegisterFilter(&OrganisationFilter{})
	s.MustRegisterFilter(&CpuCapacityFilter{})
	s.MustRegisterFilter(&RamCapacityFilter{})
	s.MustRegisterFilter(&GroupFilter{})
	s.MustRegisterFilter(&StatusFilter{})

	return s
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
func (s *Scheduler) Schedule(r *structs.Runner) error {
	if s.isShutdown {
		return nil
	}

	err := s.queue.Enqueue(r)
	if err != nil {
		return fmt.Errorf("error enqueueing runner: %w", err)
	}

	return nil
}

// MustRegisterFilter registers a Filter. If a Filter with the same name already
// exists, the program panics.
func (s *Scheduler) MustRegisterFilter(filter Filter) {
	err := s.RegisterFilter(filter)
	if err != nil {
		panic(err)
	}
}

// MustRegisterScorer registers a Scorer. If a Scorer with the same name already
// exists, the program panics.
func (s *Scheduler) MustRegisterScorer(scorer Scorer) {
	err := s.RegisterScorer(scorer)
	if err != nil {
		panic(err)
	}
}

// RegisterFilter registers a Filter. If a Filter with the same name already
// exists, an error is returned.
func (s *Scheduler) RegisterFilter(filter Filter) error {
	_, ok := s.filters[filter.Name()]
	if ok {
		return ErrFilterExists
	}

	s.filters[filter.Name()] = filter
	s.log.Debug().Msgf("registered filter %s", filter.Name())
	return nil
}

// RegisterScorer registers a Scorer. If a Scorer with the same name already
// exists, an error is returned.
func (s *Scheduler) RegisterScorer(scorer Scorer) error {
	_, ok := s.scorers[scorer.Name()]
	if ok {
		return ErrScorerExists
	}

	s.scorers[scorer.Name()] = scorer
	s.log.Debug().Msgf("registered scorer %s", scorer.Name())
	return nil
}

func (s *Scheduler) init() error {
	nodes, err := s.store.GetNodes(context.Background())
	if err != nil {
		return err
	}

	for _, n := range nodes {
		err = s.cache.AddNode(n)
		if err != nil {
			return err
		}

		s.log.Debug().Msgf("added existing node %s to internal cache", n)
	}

	runners, err := s.store.GetRunners(context.Background())
	if err != nil {
		return err
	}

	runners = runners.Filter(func(r *structs.Runner) bool {
		return r.Status == structs.RunnerStatusPending
	})

	for _, r := range runners {
		err := s.queue.Enqueue(r)
		if err != nil {
			return err
		}

		s.log.Debug().Msgf("added existing runner %s to scheduling queue", r)
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
		s.log.Error().Err(err).Msg("error dequeuing runner")
		return
	}

	cache := s.cache.DeepCopy().(Cache)

	nodes, err := cache.GetNodes()
	if err != nil {
		s.log.Error().Err(err).Msg("error getting nodes from cache")
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

	runner.Status = structs.RunnerStatusAssigned
	runner.SetNode(bestNode.Name)
	err = s.store.UpdateRunner(context.Background(), runner)
	if err != nil {
		s.log.Error().Err(err).Msg("error updating runner")
		return
	}

	cpu := int64(runner.VCPUs)
	ram := int64(runner.MemoryGB * 1024 * 1024 * 1024)
	err = s.store.ReserveNodeResources(context.Background(), bestNode.ID, cpu, ram)
	if err != nil {
		s.log.Error().Err(err).Msg("error reserving node resources")
		return
	}

	s.log.Info().Msgf("runner %s is assigned to node %s", runner.ID, bestNode.ID)
}

func findFeasibleNodes(runner *structs.Runner, nodes []*structs.Node, filters map[string]Filter) []*structs.Node {
	feasible := make([]*structs.Node, 0, len(nodes))
	for _, n := range nodes {
		if !runFilters(runner, n, filters) {
			continue
		}
		feasible = append(feasible, n)
	}

	return feasible
}

func findBestNode(runner *structs.Runner, nodes []*structs.Node, scorers map[string]Scorer) *structs.Node {
	if len(nodes) == 0 {
		return nil
	}

	nodesMap := make(map[string]*structs.Node, len(nodes))
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

	var bestNode *structs.Node
	var bestScore float64

	for nodeID, score := range scores {
		if score < bestScore {
			continue
		}

		bestNode, bestScore = nodesMap[nodeID], score
	}

	return bestNode
}

func runScorers(runner *structs.Runner, node *structs.Node, scorers map[string]Scorer) (float64, error) {
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

func runFilters(runner *structs.Runner, node *structs.Node, filters map[string]Filter) bool {
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
