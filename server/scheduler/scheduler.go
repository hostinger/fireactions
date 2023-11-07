package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler/cache"
	"github.com/hostinger/fireactions/server/scheduler/filter"
	"github.com/hostinger/fireactions/server/scheduler/filter/affinity"
	"github.com/hostinger/fireactions/server/scheduler/filter/cpucapacity"
	"github.com/hostinger/fireactions/server/scheduler/filter/heartbeat"
	"github.com/hostinger/fireactions/server/scheduler/filter/ramcapacity"
	"github.com/hostinger/fireactions/server/scheduler/filter/status"
	"github.com/hostinger/fireactions/server/scheduler/scorer"
	"github.com/hostinger/fireactions/server/scheduler/scorer/freecpu"
	"github.com/hostinger/fireactions/server/scheduler/scorer/freeram"
	"github.com/hostinger/fireactions/server/store"
	"github.com/rs/zerolog"
)

// Scheduler is responsible for scheduling Runners onto Nodes.
type Scheduler struct {
	queue        *schedulingQueue
	cache        cache.Cache
	store        store.Store
	filters      map[string]filter.Filter
	scorers      map[string]scorer.Scorer
	shutdownOnce sync.Once
	shutdownCh   chan struct{}
	isShutdown   bool
	logger       *zerolog.Logger
}

// New creates a new Scheduler.
func New(logger zerolog.Logger, store store.Store) (*Scheduler, error) {
	s := &Scheduler{
		queue:        newSchedulingQueue(),
		filters:      make(map[string]filter.Filter, 0),
		scorers:      make(map[string]scorer.Scorer, 0),
		cache:        cache.New(),
		store:        store,
		shutdownOnce: sync.Once{},
		shutdownCh:   make(chan struct{}),
		isShutdown:   false,
		logger:       &logger,
	}

	filters := []filter.Filter{cpucapacity.New(), ramcapacity.New(), heartbeat.New(), status.New(), affinity.New()}
	for _, filter := range filters {
		s.filters[filter.Name()] = filter
	}

	scorers := []scorer.Scorer{freecpu.New(), freeram.New()}
	for _, scorer := range scorers {
		s.scorers[scorer.Name()] = scorer
	}

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

// AddToQueue places a Runner into the scheduling queue. If the Runner is already
// in the queue, an error is returned.
func (s *Scheduler) AddToQueue(runners ...*fireactions.Runner) {
	if s.isShutdown {
		return
	}

	for _, r := range runners {
		s.queue.Enqueue(r)
	}
}

// RemoveFromQueue removes a Runner from the scheduling queue. If the Runner is not
// in the queue, an error is returned.
func (s *Scheduler) RemoveFromQueue(id string) error {
	if s.isShutdown {
		return nil
	}

	s.queue.Remove(id)
	return nil
}

func (s *Scheduler) init() error {
	nodes, err := s.store.GetNodes(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		err = s.cache.AddNode(n)
		if err != nil {
			return err
		}

		s.logger.Debug().Str("node", n.Name).Msgf("added existing node to scheduler cache")
	}

	runners, err := s.store.GetRunners(context.Background(), func(m *fireactions.Runner) bool {
		return m.NodeID == nil && m.DeletedAt == nil
	})
	if err != nil {
		return err
	}

	for _, r := range runners {
		s.queue.Enqueue(r)
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
	if err != nil {
		s.logger.Error().Err(err).Msg("error dequeuing runner")
		return
	}

	cache := s.cache.DeepCopy().(cache.Cache)

	nodes, err := cache.GetNodes()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting nodes from cache")
		return
	}

	feasibleNodes := s.findFeasibleNodes(runner, nodes)
	if len(feasibleNodes) == 0 {
		s.queue.Block(runner.ID)
		return
	}

	bestNode := s.findBestNode(feasibleNodes)
	if bestNode == nil {
		s.queue.Block(runner.ID)
		return
	}

	err = s.store.AllocateRunner(context.Background(), bestNode.ID, runner.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("error assigning runner to node")
		return
	}

	s.logger.Info().Str("runner", runner.Name).Str("node", bestNode.Name).
		Msg("assigned runner to node")
}

func (s *Scheduler) findFeasibleNodes(runner *fireactions.Runner, nodes []*fireactions.Node) []*fireactions.Node {
	feasible := make([]*fireactions.Node, 0, len(nodes))

	results := make(map[string]error, len(nodes))
	for _, node := range nodes {
		ok, reason := s.runFilters(runner, node)
		if !ok {
			results[node.ID] = reason
			continue
		}

		feasible = append(feasible, node)
	}

	if len(feasible) == 0 {
		s.logger.Debug().Str("runner", runner.Name).Msgf("scheduler: no feasible nodes found for runner: %d/%d nodes filtered out: %v", len(nodes)-len(feasible), len(nodes), results)
	}

	return feasible
}

func (s *Scheduler) runFilters(runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	for _, filter := range s.filters {
		ok, reason := filter.Filter(context.Background(), runner, node)

		if !ok {
			return false, reason
		}

		continue
	}

	return true, nil
}

func (s *Scheduler) findBestNode(nodes []*fireactions.Node) *fireactions.Node {
	nodesMap := make(map[string]*fireactions.Node, len(nodes))
	for _, n := range nodes {
		nodesMap[n.ID] = n
	}

	scores := make(map[string]float64, len(nodes))
	for _, node := range nodes {
		result, err := s.runScorers(node)
		if err != nil {
			continue
		}

		scores[node.ID] = result
	}

	var bestNode *fireactions.Node
	var bestScore float64

	for nodeID, score := range scores {
		if score < bestScore {
			continue
		}

		bestNode, bestScore = nodesMap[nodeID], score
	}

	return bestNode
}

func (s *Scheduler) runScorers(node *fireactions.Node) (float64, error) {
	var score float64
	for _, scorer := range s.scorers {
		result, err := scorer.Score(node)
		if err != nil {
			return 0, err
		}

		score += result
	}

	return score, nil
}
