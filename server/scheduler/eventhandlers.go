package scheduler

import "github.com/hostinger/fireactions"

func (s *Scheduler) NotifyNodeCreated(n *fireactions.Node) {
	s.handleNodeCreated(n)
}

func (s *Scheduler) NotifyNodeUpdated(n *fireactions.Node) {
	s.handleNodeUpdated(n)
}

func (s *Scheduler) NotifyNodeDeleted(n *fireactions.Node) {
	s.handleNodeDeleted(n)
}

func (s *Scheduler) handleNodeCreated(node *fireactions.Node) {
	err := s.cache.AddNode(node)
	if err != nil {
		s.logger.Error().Err(err).Msg("error adding node to cache")
		return
	}

	s.queue.UnblockAll()
}

func (s *Scheduler) handleNodeUpdated(node *fireactions.Node) {
	err := s.cache.PutNode(node)
	if err != nil {
		s.logger.Error().Err(err).Msg("error updating node in cache")
		return
	}

	s.queue.UnblockAll()
}

func (s *Scheduler) handleNodeDeleted(node *fireactions.Node) {
	err := s.cache.DelNode(node)
	if err != nil {
		s.logger.Error().Err(err).Msg("error deleting node from cache")
		return
	}

	s.queue.UnblockAll()
}
