package printer

import "github.com/hostinger/fireactions/api"

// Runner is a Printable for api.Runners
type Runner struct {
	Runners api.Runners
}

var _ Printable = &Runner{}

// Cols returns the columns for the Printable
func (r *Runner) Cols() []string {
	cols := []string{
		"Name", "Node", "Organisation", "Group", "Status", "Flavor", "Created At", "Updated At",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (r *Runner) ColsMap() map[string]string {
	cols := map[string]string{
		"Name": "Name", "Node": "Node", "Organisation": "Organisation", "Group": "Group", "Status": "Status", "Flavor": "Flavor", "Created At": "Created At", "Updated At": "Updated At",
	}

	return cols
}

// KV returns the key value for the Printable
func (r *Runner) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(r.Runners))
	for _, runner := range r.Runners {
		node := "N/A (Unassigned)"
		if runner.Node != nil {
			node = *runner.Node
		}

		createdAt := runner.CreatedAt.Format("2006-01-02 15:04:05")
		updatedAt := runner.UpdatedAt.Format("2006-01-02 15:04:05")

		kv = append(kv, map[string]interface{}{
			"Name": runner.Name, "Node": node, "Organisation": runner.Organisation, "Group": runner.Group, "Status": runner.Status, "Flavor": runner.Flavor.Name, "Created At": createdAt, "Updated At": updatedAt,
		})
	}

	return kv
}
