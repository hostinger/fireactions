package printer

import (
	"fmt"

	"github.com/hostinger/fireactions/api"
)

// Group is a Printable for api.Groups
type Group struct {
	Groups api.Groups
}

var _ Printable = &Group{}

// Cols returns the columns for the Printable
func (g *Group) Cols() []string {
	cols := []string{
		"Name", "Enabled", "Is Default",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (g *Group) ColsMap() map[string]string {
	cols := map[string]string{
		"Name": "Name", "Enabled": "Enabled", "Is Default": "IsDefault",
	}

	return cols
}

// KV returns the key value for the Printable
func (g *Group) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(g.Groups))
	for _, group := range g.Groups {
		kv = append(kv, map[string]interface{}{
			"Name": group.Name, "Enabled": fmt.Sprintf("%t", group.Enabled), "Is Default": fmt.Sprintf("%t", group.IsDefault),
		})
	}

	return kv
}
