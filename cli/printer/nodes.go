package printer

import (
	"fmt"
	"strings"

	"github.com/hostinger/fireactions/api"
)

// Node is a Printable for api.Nodes
type Node struct {
	Nodes api.Nodes
}

var _ Printable = &Node{}

// Cols returns the columns for the Printable
func (n *Node) Cols() []string {
	cols := []string{
		"ID", "Name", "Is Cordoned", "Organisation", "Status", "Groups", "Cpu Usage", "Mem Usage", "Last Seen",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (n *Node) ColsMap() map[string]string {
	cols := map[string]string{
		"ID": "ID", "Name": "Name", "Is Cordoned": "IsCordoned", "Organisation": "Organisation", "Status": "Status", "Groups": "Groups", "Cpu Usage": "CpuUsage", "Mem Usage": "MemUsage", "Last Seen": "LastSeen",
	}

	return cols
}

// KV returns the key value for the Printable
func (n *Node) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(n.Nodes))
	for _, node := range n.Nodes {
		cpuUsage := fmt.Sprintf("%d/%d Cores (%.0f%%)", node.CpuTotal-node.CpuFree, node.CpuTotal, float64(node.CpuTotal-node.CpuFree)/float64(node.CpuTotal)*100)
		memUsage := fmt.Sprintf("%.2f/%.2f GB (%.0f%%)", float64(node.MemTotal-node.MemFree)/1024/1024/1024,
			float64(node.MemTotal)/1024/1024/1024, float64(node.MemTotal-node.MemFree)/float64(node.MemTotal)*100)

		kv = append(kv, map[string]interface{}{
			"ID": node.ID, "Name": node.Name, "Is Cordoned": fmt.Sprintf("%t", node.IsCordoned), "Organisation": node.Organisation, "Status": node.Status,
			"Groups": strings.Join(node.Groups, ","), "Cpu Usage": cpuUsage, "Mem Usage": memUsage, "Last Seen": node.LastSeen.Format("2006-01-02 15:04:05"),
		})
	}

	return kv
}
