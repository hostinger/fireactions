package structs

import (
	"fmt"
)

// Flavor struct.
type Flavor struct {
	Name         string `mapstructure:"name"`
	DiskSizeGB   int64  `mapstructure:"disk"`
	MemorySizeMB int64  `mapstructure:"mem"`
	VCPUs        int64  `mapstructure:"cpu"`
	ImageName    string `mapstructure:"image"`
}

// String returns a string representation of a Flavor.
func (f *Flavor) String() string {
	return fmt.Sprintf("%s (vCPUs: %d, Memory: %d MB, Disk: %d GB)", f.Name, f.VCPUs, f.MemorySizeMB, f.DiskSizeGB)
}

// GetMemorySizeBytes returns the memory size in bytes.
func (f *Flavor) GetMemorySizeBytes() int64 {
	return f.MemorySizeMB * 1024 * 1024
}
