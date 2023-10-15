package models

import (
	"fmt"
)

// Flavor struct.
type Flavor struct {
	Name         string
	Enabled      bool
	DiskSizeGB   int64
	MemorySizeMB int64
	VCPUs        int64
	Image        string
}

// String returns a string representation of a Flavor.
func (f *Flavor) String() string {
	return fmt.Sprintf("%s (Enabled: %t, vCPUs: %d, Memory: %dMB, Disk: %dGB, Image: %s)", f.Name, f.Enabled, f.VCPUs, f.MemorySizeMB, f.DiskSizeGB, f.Image)
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
