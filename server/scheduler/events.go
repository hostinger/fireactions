package scheduler

import "github.com/hostinger/fireactions/server/models"

func (s *Scheduler) HandleEvent(e *models.Event) {
	switch e.Topic {
	case models.EventTopicNode:
		s.handleNodeEvent(*e)
	default:
		s.logger.Warn().Msgf("unknown event topic: %s", e.Topic)
	}
}

func (s *Scheduler) handleNodeEvent(e models.Event) {
	switch e.Type {
	case models.EventTypeNodeCreated:
		s.onNodeCreated(e.Object.(*models.Node))
	case models.EventTypeNodeUpdated:
		s.onNodeUpdated(e.Object.(*models.Node))
	case models.EventTypeNodeDeleted:
		s.onNodeDeleted(e.Object.(*models.Node))
	}
}

func (s *Scheduler) onNodeCreated(node *models.Node) {
	err := s.cache.AddNode(node)
	if err != nil {
		s.logger.Error().Err(err).Msg("error adding node to cache")
		return
	}

	s.queue.UnblockAll()
}

func (s *Scheduler) onNodeUpdated(node *models.Node) {
	err := s.cache.PutNode(node)
	if err != nil {
		s.logger.Error().Err(err).Msg("error updating node in cache")
		return
	}

	s.queue.UnblockAll()
}

func (s *Scheduler) onNodeDeleted(node *models.Node) {
	err := s.cache.DelNode(node)
	if err != nil {
		s.logger.Error().Err(err).Msg("error deleting node from cache")
		return
	}

	s.queue.UnblockAll()
}
