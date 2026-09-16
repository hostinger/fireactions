package server

import (
	"testing"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("testdata/config1.yaml")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, "testdata/config1.yaml", config.path)
}

func TestNewConfigNetworkInterface(t *testing.T) {
	config, err := NewConfig("testdata/config1.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	networkInterface := config.Pools[0].Firecracker.NetworkInterface
	if networkInterface == nil {
		t.Fatal("expected network interface configuration to be set")
	}

	assert.Equal(t, &FirecrackerTokenBucketConfig{
		Size: 131072000, RefillTime: 1000, OneTimeBurst: firecracker.Int64(262144000),
	}, networkInterface.InRateLimiter.Bandwidth)
	assert.Equal(t, &FirecrackerTokenBucketConfig{
		Size: 10000, RefillTime: 1000,
	}, networkInterface.InRateLimiter.Ops)
	assert.Equal(t, &FirecrackerTokenBucketConfig{
		Size: 26214400, RefillTime: 1000,
	}, networkInterface.OutRateLimiter.Bandwidth)
	assert.Nil(t, networkInterface.OutRateLimiter.Ops)

	// A pool without a network_interface block leaves the interface unlimited.
	assert.Nil(t, config.Pools[1].Firecracker.NetworkInterface)
}

func TestNewConfigNetworkInterfaceInvalid(t *testing.T) {
	// A token bucket without a size is rejected.
	_, err := NewConfig("testdata/config2.yaml")
	assert.ErrorContains(t, err, "Config.Pools[0].Firecracker.NetworkInterface.InRateLimiter.Bandwidth.Size")
}

func TestNewConfigRootfs(t *testing.T) {
	config, err := NewConfig("testdata/config1.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rootfs := config.Pools[0].Firecracker.Rootfs
	if rootfs == nil {
		t.Fatal("expected rootfs configuration to be set")
	}

	assert.Equal(t, &FirecrackerTokenBucketConfig{
		Size: 52428800, RefillTime: 1000,
	}, rootfs.RateLimiter.Bandwidth)
	assert.Equal(t, &FirecrackerTokenBucketConfig{
		Size: 2000, RefillTime: 1000,
	}, rootfs.RateLimiter.Ops)

	// A pool without a rootfs block leaves the block device unlimited.
	assert.Nil(t, config.Pools[1].Firecracker.Rootfs)
}

func TestNewConfigRootfsInvalid(t *testing.T) {
	// A token bucket without a refill time is rejected.
	_, err := NewConfig("testdata/config3.yaml")
	assert.ErrorContains(t, err, "Config.Pools[0].Firecracker.Rootfs.RateLimiter.Ops.RefillTime")
}

func TestFirecrackerRateLimiterConfigToSDK(t *testing.T) {
	var nilRateLimiter *FirecrackerRateLimiterConfig
	assert.Nil(t, nilRateLimiter.toSDK())

	rateLimiter := &FirecrackerRateLimiterConfig{
		Bandwidth: &FirecrackerTokenBucketConfig{Size: 131072000, RefillTime: 1000, OneTimeBurst: firecracker.Int64(262144000)},
	}

	assert.Equal(t, &models.RateLimiter{
		Bandwidth: &models.TokenBucket{
			Size:         firecracker.Int64(131072000),
			RefillTime:   firecracker.Int64(1000),
			OneTimeBurst: firecracker.Int64(262144000),
		},
	}, rateLimiter.toSDK())
}
