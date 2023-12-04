package scheduler

import (
	"context"
	"fmt"

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
	queue   *schedulingQueue
	cache   cache.Cache
	store   store.Store
	filters map[string]filter.Filter
	scorers map[string]scorer.Scorer
	logger  *zerolog.Logger
}

// New creates a new Scheduler.
func New(logger zerolog.Logger, store store.Store) (*Scheduler, error) {
	s := &Scheduler{
		queue:   newSchedulingQueue(),
		filters: make(map[string]filter.Filter, 0),
		scorers: make(map[string]scorer.Scorer, 0),
		cache:   cache.New(),
		store:   store,
		logger:  &logger,
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
func (s *Scheduler) Run(ctx context.Context) error {
	err := s.init(ctx)
	if err != nil {
		return fmt.Errorf("error initializing scheduler: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			s.schedule()
		}
	}()

	return nil
}

// AddToQueue places a Runner into the scheduling queue. If the Runner is already
// in the queue, an error is returned.
func (s *Scheduler) AddToQueue(runner *fireactions.Runner) {
	s.queue.Enqueue(runner)
}

func (s *Scheduler) init(ctx context.Context) error {
	nodes, err := s.store.GetNodes(ctx, nil)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		err = s.cache.AddNode(n)
		if err != nil {
			return err
		}
	}

	runners, err := s.store.GetRunners(ctx, func(m *fireactions.Runner) bool {
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

	tx, err := s.store.BeginTransaction()
	if err != nil {
		s.logger.Error().Err(err).Msg("error beginning transaction")
		return
	}
	defer tx.Rollback()

	_, err = s.store.UpdateRunnerWithTransaction(context.Background(), tx, runner.ID, func(r *fireactions.Runner) error {
		r.NodeID = &bestNode.ID
		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("error updating runner")
		return
	}

	node, err := s.store.UpdateNodeWithTransaction(context.Background(), tx, bestNode.ID, func(n *fireactions.Node) error {
		n.CPU.Reserve(runner.Resources.VCPUs)
		n.RAM.Reserve(runner.Resources.MemoryMB * 1024 * 1024)
		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("error updating node")
		return
	}

	err = tx.Commit()
	if err != nil {
		s.logger.Error().Err(err).Msg("error committing transaction")
		return
	}

	s.cache.PutNode(node)
	s.logger.Debug().Msgf("scheduler: assigned runner %s to node %s", runner.Name, bestNode.Name)
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
		s.logger.Debug().Msgf("scheduler: no feasible nodes found for runner %s: %d/%d nodes filtered out: %v", runner.Name, len(nodes)-len(feasible), len(nodes), results)
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
