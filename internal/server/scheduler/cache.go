package scheduler

import (
	"github.com/hostinger/fireactions/internal/structs"
)

type Cache interface {
	GetNodes() ([]*structs.Node, error)
	AddNode(n *structs.Node) error
	DelNode(n *structs.Node) error
	PutNode(n *structs.Node) error
	DeepCopy() interface{}
}

// AddNodeToCache adds a node to the internal cache of the Scheduler. This
// also unblocks any unschedulable Runners in the scheduling queue.
func (s *Scheduler) AddNodeToCache(node *structs.Node) {
	err := s.cache.AddNode(node)
	if err != nil {
		s.log.Error().Err(err).Msg("error adding node to cache")
		return
	}

	s.queue.UnblockAll()
}

// DeleteNodeFromCache deletes a node from the internal cache of the Scheduler.
// This also unblocks any unschedulable Runners in the scheduling queue.
func (s *Scheduler) DeleteNodeFromCache(node *structs.Node) {
	err := s.cache.DelNode(node)
	if err != nil {
		s.log.Error().Err(err).Msg("error deleting node from cache")
		return
	}

	s.queue.UnblockAll()
}

// UpdateNodeInCache updates a node in the internal cache of the Scheduler.
// This also unblocks any unschedulable Runners in the scheduling queue.
func (s *Scheduler) UpdateNodeInCache(node *structs.Node) {
	err := s.cache.PutNode(node)
	if err != nil {
		s.log.Error().Err(err).Msg("error updating node in cache")
		return
	}

	s.queue.UnblockAll()
}
