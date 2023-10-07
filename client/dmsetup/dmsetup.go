package dmsetup

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrDeviceAlreadyExists = fmt.Errorf("device already exists")
)

var client *Client

// Client is a wrapper around the dmsetup binary.
type Client struct {
	BinaryPath string
}

// DefaultClient returns the default client with BinaryPath set to /usr/sbin/dmsetup.
func DefaultClient() *Client {
	if client == nil {
		client = NewClient("/usr/sbin/dmsetup")
	}

	return client
}

// NewClient returns a new client with the given binary path.
func NewClient(binaryPath string) *Client {
	if binaryPath == "" {
		binaryPath = "/usr/sbin/dmsetup"
	}

	c := &Client{
		BinaryPath: binaryPath,
	}

	return c
}

// Create creates a new device with the given name, path and table.
func (c *Client) Create(ctx context.Context, deviceName, table string) error {
	if _, err := c.dmsetup(ctx, "info", deviceName); err == nil {
		return ErrDeviceAlreadyExists
	}

	_, err := c.dmsetup(ctx, "create", deviceName, "--table", table)
	return err
}

// Remove removes the given device.
func (c *Client) Remove(ctx context.Context, deviceName string) error {
	_, err := c.dmsetup(ctx, "remove", deviceName)
	return err
}

func (c *Client) dmsetup(ctx context.Context, args ...string) (string, error) {
	data, err := exec.CommandContext(ctx, c.BinaryPath, args...).CombinedOutput()
	output := string(data)
	if err != nil {
		return "", fmt.Errorf("dmsetup %s\nerror: %s\n: %w", strings.Join(args, " "), output, err)
	}

	output = strings.TrimSuffix(output, "\n")
	output = strings.TrimSpace(output)

	return output, nil
}
