package printer

import "github.com/hostinger/fireactions/api"

// Node is a Printable for api.Nodes
type Node struct {
	Nodes api.Nodes
}

var _ Printable = &Node{}

// Cols returns the columns for the Printable
func (n *Node) Cols() []string {
	cols := []string{
		"ID", "Name", "Organisation", "Status", "Group", "CpuTotal", "CpuFree", "MemTotal", "MemFree", "LastSeen",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (n *Node) ColsMap() map[string]string {
	cols := map[string]string{
		"ID": "ID", "Name": "Name", "Organisation": "Organisation", "Status": "Status", "Group": "Group", "CpuTotal": "CpuTotal", "CpuFree": "CpuFree", "MemTotal": "MemTotal", "MemFree": "MemFree", "LastSeen": "LastSeen",
	}

	return cols
}

// KV returns the key value for the Printable
func (n *Node) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(n.Nodes))
	for _, node := range n.Nodes {
		kv = append(kv, map[string]interface{}{
			"ID": node.ID, "Name": node.Name, "Organisation": node.Organisation, "Status": node.Status, "Group": node.Group, "CpuTotal": node.CpuTotal, "CpuFree": node.CpuFree, "MemTotal": node.MemTotal, "MemFree": node.MemFree, "LastSeen": node.LastSeen,
		})
	}

	return kv
}
