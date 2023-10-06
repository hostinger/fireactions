package structs

import (
	"fmt"
)

// Flavor struct.
type Flavor struct {
	Name         string
	DiskSizeGB   int64
	MemorySizeMB int64
	VCPUs        int64
	ImageName    string
	Enabled      bool
}

// String returns a string representation of a Flavor.
func (f *Flavor) String() string {
	return fmt.Sprintf("%s (vCPUs: %d, Memory: %d MB, Disk: %d GB, Enabled: %t)", f.Name, f.VCPUs, f.MemorySizeMB, f.DiskSizeGB, f.Enabled)
}

// GetMemorySizeBytes returns the memory size in bytes.
func (f *Flavor) GetMemorySizeBytes() int64 {
	return f.MemorySizeMB * 1024 * 1024
}

// Enable enables the Flavor for usage.
func (f *Flavor) Enable() {
	f.Enabled = true
}

// Disable disables the Flavor for usage.
func (f *Flavor) Disable() {
	f.Enabled = false
}
