package printer

import (
	"fmt"

	"github.com/hostinger/fireactions/api"
)

// Flavor is a Printable for api.Flavors
type Flavor struct {
	Flavors api.Flavors
}

var _ Printable = &Flavor{}

// Cols returns the columns for the Printable
func (f *Flavor) Cols() []string {
	cols := []string{
		"Name", "Enabled", "Is Default", "VCPUs", "Memory", "Disk", "Image",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (f *Flavor) ColsMap() map[string]string {
	cols := map[string]string{
		"Name": "Name", "Enabled": "Enabled", "Is Default": "IsDefault", "VCPUs": "VCPUs", "Memory": "Memory", "Disk": "Disk", "Image": "Image",
	}

	return cols
}

// KV returns the key value for the Printable
func (f *Flavor) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(f.Flavors))
	for _, flavor := range f.Flavors {
		kv = append(kv, map[string]interface{}{
			"Name": flavor.Name, "Enabled": fmt.Sprintf("%t", flavor.Enabled), "Is Default": fmt.Sprintf("%t", flavor.IsDefault), "VCPUs": flavor.VCPUs, "Memory": flavor.MemorySizeMB, "Disk": flavor.DiskSizeGB, "Image": flavor.Image,
		})
	}

	return kv
}
