package cache

import (
	"sync"

	"github.com/hostinger/fireactions"
)

type Cache interface {
	GetNodes() ([]*fireactions.Node, error)
	AddNode(n *fireactions.Node) error
	DelNode(n *fireactions.Node) error
	PutNode(n *fireactions.Node) error
	DeepCopy() interface{}
}

type cacheImpl struct {
	nodes   map[string]*fireactions.Node
	nodesMu sync.RWMutex
}

// New creates a new implementation of Scheduler Cache
func New() *cacheImpl {
	c := &cacheImpl{
		nodes:   make(map[string]*fireactions.Node),
		nodesMu: sync.RWMutex{},
	}

	return c
}

// AddNode adds a new node to the cache
func (c *cacheImpl) AddNode(node *fireactions.Node) error {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()

	_, ok := c.nodes[node.Name]
	if ok {
		return nil
	}

	c.nodes[node.Name] = node
	return nil
}

// PutNode updates a node in the cache
func (c *cacheImpl) PutNode(node *fireactions.Node) error {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()

	_, ok := c.nodes[node.Name]
	if !ok {
		return nil
	}

	c.nodes[node.Name] = node
	return nil
}

// DelNode deletes a node from the cache
func (c *cacheImpl) DelNode(node *fireactions.Node) error {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()

	delete(c.nodes, node.Name)
	return nil
}

// GetNode gets a node from the cache
func (c *cacheImpl) GetNodes() ([]*fireactions.Node, error) {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()

	nodes := make([]*fireactions.Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// DeepCopy creates a deep copy of the cache object
func (c *cacheImpl) DeepCopy() interface{} {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()

	nodes := make(map[string]*fireactions.Node, len(c.nodes))
	for k, v := range c.nodes {
		nodes[k] = v
	}

	return &cacheImpl{nodesMu: sync.RWMutex{}, nodes: nodes}
}
