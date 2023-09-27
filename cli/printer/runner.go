package printer

import (
	"fmt"
	"strings"

	timeago "github.com/caarlos0/timea.go"
	"github.com/hostinger/fireactions"
)

type Runner struct {
	Runners []*fireactions.Runner
}

var _ Printable = &Runner{}

func (r *Runner) Cols() []string {
	cols := []string{
		"ID", "Name", "Phase", "Organisation", "Node", "Labels", "CPU", "RAM", "Created", "Updated",
	}

	return cols
}

func (r *Runner) ColsMap() map[string]string {
	cols := map[string]string{
		"ID":           "ID",
		"Name":         "Name",
		"Phase":        "Phase",
		"Organisation": "Organisation",
		"Node":         "Node",
		"Labels":       "Labels",
		"CPU":          "CPU",
		"RAM":          "RAM",
		"Created":      "Created",
		"Updated":      "Updated",
	}

	return cols
}

func (r *Runner) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(r.Runners))
	for _, runner := range r.Runners {
		cpu := fmt.Sprintf("%d VCPUs", runner.Resources.VCPUs)
		ram := fmt.Sprintf("%d MB", runner.Resources.MemoryBytes/1024/1024)
		created := timeago.Of(runner.CreatedAt)
		updated := timeago.Of(runner.UpdatedAt)
		labels := strings.Join(runner.Labels, ",")
		var node string
		if runner.NodeID == nil {
			node = "N/A"
		} else {
			node = *runner.NodeID
		}

		kv = append(kv, map[string]interface{}{
			"ID":           runner.ID,
			"Name":         runner.Name,
			"Phase":        runner.Status.Phase,
			"Organisation": runner.Organisation,
			"Node":         node,
			"Labels":       labels,
			"CPU":          cpu,
			"RAM":          ram,
			"Created":      created,
			"Updated":      updated,
		})
	}

	return kv
}
