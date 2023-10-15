package cache

import (
	"sync"

	"github.com/hostinger/fireactions/server/models"
)

type Cache interface {
	GetNodes() ([]*models.Node, error)
	AddNode(n *models.Node) error
	DelNode(n *models.Node) error
	PutNode(n *models.Node) error
	DeepCopy() interface{}
}

type cacheImpl struct {
	nodes   map[string]*models.Node
	nodesMu sync.RWMutex
}

// New creates a new implementation of Scheduler Cache
func New() *cacheImpl {
	c := &cacheImpl{
		nodes:   make(map[string]*models.Node),
		nodesMu: sync.RWMutex{},
	}

	return c
}

// AddNode adds a new node to the cache
func (c *cacheImpl) AddNode(node *models.Node) error {
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
func (c *cacheImpl) PutNode(node *models.Node) error {
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
func (c *cacheImpl) DelNode(node *models.Node) error {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()

	delete(c.nodes, node.Name)
	return nil
}

// GetNode gets a node from the cache
func (c *cacheImpl) GetNodes() ([]*models.Node, error) {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()

	nodes := make([]*models.Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// DeepCopy creates a deep copy of the cache object
func (c *cacheImpl) DeepCopy() interface{} {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()

	nodes := make(map[string]*models.Node, len(c.nodes))
	for k, v := range c.nodes {
		nodes[k] = v
	}

	return &cacheImpl{nodesMu: sync.RWMutex{}, nodes: nodes}
}
