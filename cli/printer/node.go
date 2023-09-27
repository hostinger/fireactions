package printer

import (
	"fmt"

	timeago "github.com/caarlos0/timea.go"
	"github.com/hostinger/fireactions"
)

// Node is a Printable for api.Nodes
type Node struct {
	Nodes []*fireactions.Node
}

var _ Printable = &Node{}

// Cols returns the columns for the Printable
func (n *Node) Cols() []string {
	cols := []string{
		"ID", "Name", "Status", "Cpu Usage", "Mem Usage", "Last Reconcile", "Created", "Updated",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (n *Node) ColsMap() map[string]string {
	cols := map[string]string{
		"ID": "ID", "Name": "Name", "Status": "Status", "State": "State", "Region": "Region", "Cpu Usage": "Cpu Usage", "Mem Usage": "Mem Usage",
		"Last Reconcile": "Last Reconcile", "Created": "Created", "Updated": "Updated",
	}

	return cols
}

// KV returns the key value for the Printable
func (n *Node) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(n.Nodes))
	for _, node := range n.Nodes {
		cpuUsage := fmt.Sprintf("%d/%d Cores (%.0f%%)", node.CPU.Allocated, node.CPU.Capacity, float64(node.CPU.Allocated)/float64(node.CPU.Capacity)*100)
		memUsage := fmt.Sprintf("%.2f/%.2f GB (%.0f%%)", float64(node.RAM.Allocated)/1024/1024/1024,
			float64(node.RAM.Capacity)/1024/1024/1024, float64(node.RAM.Allocated)/float64(node.RAM.Capacity)*100)
		lastReconcile := timeago.Of(node.LastReconcileAt)
		created := timeago.Of(node.CreatedAt)
		updated := timeago.Of(node.UpdatedAt)

		kv = append(kv, map[string]interface{}{
			"ID":             node.ID,
			"Name":           node.Name,
			"Status":         node.Status,
			"Cpu Usage":      cpuUsage,
			"Mem Usage":      memUsage,
			"Last Reconcile": lastReconcile,
			"Created":        created,
			"Updated":        updated,
		})
	}

	return kv
}
