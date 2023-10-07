package scheduler

import "github.com/hostinger/fireactions/server/structs"

func (s *Scheduler) HandleEvent(e *structs.Event) {
	switch e.Topic {
	case structs.EventTopicNode:
		s.handleNodeEvent(*e)
	default:
		s.log.Warn().Msgf("unknown event topic: %s", e.Topic)
	}
}

func (s *Scheduler) handleNodeEvent(e structs.Event) {
	switch e.Type {
	case structs.EventTypeNodeCreated:
		s.onNodeCreated(e.Object.(*structs.Node))
	case structs.EventTypeNodeUpdated:
		s.onNodeUpdated(e.Object.(*structs.Node))
	case structs.EventTypeNodeDeleted:
		s.onNodeDeleted(e.Object.(*structs.Node))
	}
}

func (s *Scheduler) onNodeCreated(node *structs.Node) {
	err := s.cache.AddNode(node)
	if err != nil {
		s.log.Error().Err(err).Msg("error adding node to cache")
		return
	}

	s.queue.UnblockAll()
}

func (s *Scheduler) onNodeUpdated(node *structs.Node) {
	err := s.cache.PutNode(node)
	if err != nil {
		s.log.Error().Err(err).Msg("error updating node in cache")
		return
	}

	s.queue.UnblockAll()
}

func (s *Scheduler) onNodeDeleted(node *structs.Node) {
	err := s.cache.DelNode(node)
	if err != nil {
		s.log.Error().Err(err).Msg("error deleting node from cache")
		return
	}

	s.queue.UnblockAll()
}
