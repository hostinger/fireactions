package firecracker

import (
	"net"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

type Config struct {
	*firecracker.Config

	Metadata map[string]interface{}
}

type configBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new ConfigBuilder for building a firecracker.Config.
func NewConfigBuilder() *configBuilder {
	metadata := map[string]interface{}{"latest": map[string]interface{}{"meta-data": map[string]interface{}{}}}
	b := &configBuilder{config: &Config{
		Metadata: metadata,
		Config:   &firecracker.Config{},
	}}

	return b
}

// WithID sets the VMID.
func (b *configBuilder) WithID(id string) *configBuilder {
	b.config.VMID = id
	return b
}

// WithVCPUs sets the number of VCPUs for the virtual machine.
func (b *configBuilder) WithVCPUs(vcpus int64) *configBuilder {
	b.config.MachineCfg.VcpuCount = &vcpus
	return b
}

// WithMemoryMB sets the amount of memory in MB for the virtual machine.
func (b *configBuilder) WithMemoryMB(memoryMB int64) *configBuilder {
	b.config.MachineCfg.MemSizeMib = firecracker.Int64(memoryMB)
	return b
}

// WithMemory sets the amount of memory in bytes for the virtual machine.
func (b *configBuilder) WithMemory(memoryBytes int64) *configBuilder {
	b.config.MachineCfg.MemSizeMib = firecracker.Int64(memoryBytes / 1024 / 1024)
	return b
}

func (b *configBuilder) WithSocketPath(path string) *configBuilder {
	b.config.SocketPath = path
	return b
}

// WithKernelImagePath sets the path to the kernel image.
func (b *configBuilder) WithKernelImagePath(path string) *configBuilder {
	b.config.KernelImagePath = path
	return b
}

// WithKernelArgs sets the kernel arguments.
func (b *configBuilder) WithKernelArgs(args string) *configBuilder {
	b.config.KernelArgs = args
	return b
}

// WithLogLevel sets the log level.
func (b *configBuilder) WithLogLevel(level string) *configBuilder {
	b.config.LogLevel = level
	return b
}

// WithLogPath sets the path to the log file.
func (b *configBuilder) WithLogPath(path string) *configBuilder {
	b.config.LogPath = path
	return b
}

// WithCNINetworkInterface sets a CNI network interface for the virtual machine.
func (b *configBuilder) WithCNINetworkInterface(networkName, IfName, confDir string, binPaths []string, allowMMDS bool) *configBuilder {
	b.config.NetworkInterfaces = append(b.config.NetworkInterfaces, firecracker.NetworkInterface{
		AllowMMDS: allowMMDS,
		CNIConfiguration: &firecracker.CNIConfiguration{
			NetworkName: networkName,
			IfName:      IfName,
			BinPath:     binPaths,
			ConfDir:     confDir,
		},
		InRateLimiter:  &models.RateLimiter{},
		OutRateLimiter: &models.RateLimiter{},
	})

	return b
}

// WithDrive sets a drive for the virtual machine.
func (b *configBuilder) WithDrive(id, pathOnHost string, isRootDevice, isReadonly bool) *configBuilder {
	b.config.Drives = append(b.config.Drives, models.Drive{
		DriveID:      &id,
		PathOnHost:   &pathOnHost,
		IsRootDevice: &isRootDevice,
		IsReadOnly:   &isReadonly,
	})

	return b
}

// WithMetadata sets a metadata key/value pair for the MMDS.
func (b *configBuilder) WithMetadata(key, value interface{}) *configBuilder {
	b.config.Metadata["latest"].(map[string]interface{})["meta-data"].(map[string]interface{})[key.(string)] = value
	return b
}

// WithMMDSVersion sets the version for the MMDS.
func (b *configBuilder) WithMMDSVersion(version string) *configBuilder {
	b.config.MmdsVersion = firecracker.MMDSVersion(version)
	return b
}

// WithMMDSAddress sets the address for the MMDS.
func (b *configBuilder) WithMMDSAddress(address net.IP) *configBuilder {
	b.config.MmdsAddress = address
	return b
}

// Build validates the configuration and returns a firecracker.Config.
func (b *configBuilder) Build() (*Config, error) {
	err := b.config.Validate()
	if err != nil {
		return nil, err
	}

	return b.config, nil
}
