package structs

import (
	"fmt"
	"time"
)

// NodeFilterFunc is a function that filters Nodes. If the function returns true, the Node is included in the
// result, otherwise it is excluded.
type NodeFilterFunc func(*Node) bool

// Nodes is a slice of Node.
type Nodes []*Node

// Node struct.
type Node struct {
	ID           string
	Name         string
	Organisation string
	Group        string
	Status       NodeStatus
	CPU          Resource
	RAM          Resource
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NodeStatus represents the status of a Node.
type NodeStatus string

const (
	NodeStatusUnknown NodeStatus = "Unknown"
	NodeStatusOnline  NodeStatus = "Online"
	NodeStatusOffline NodeStatus = "Offline"
)

// String returns a string representation of a Node.
func (n *Node) String() string {
	return fmt.Sprintf("%s (%s)", n.Name, n.ID)
}

// Filter filters Nodes using a NodeFilterFunc.
func (n Nodes) Filter(f NodeFilterFunc) Nodes {
	var nodes Nodes
	for _, node := range n {
		if !f(node) {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
}
