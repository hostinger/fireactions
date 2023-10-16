package models

import (
	"fmt"
	"time"
)

// NodeFilterFunc is a function that filters Nodes. If the function returns true, the Node is included in the
// result, otherwise it is excluded.
type NodeFilterFunc func(*Node) bool

// Node struct.
type Node struct {
	ID           string
	Name         string
	Organisation string
	Groups       []*Group
	Status       NodeStatus
	CPU          Resource
	RAM          Resource
	IsCordoned   bool
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

// FilterNodes filters a slice of Nodes using a NodeFilterFunc. If the function returns true, the Node is
// included in the result, otherwise it is excluded.
func FilterNodes(nodes []*Node, filter NodeFilterFunc) []*Node {
	filtered := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		if filter(node) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}
