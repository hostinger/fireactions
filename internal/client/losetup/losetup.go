package losetup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var client *Client

// Client is an implementation of the 'losetup' command
type Client struct {
	BinaryPath string
}

// DefaultClient returns the default client with BinaryPath set to /usr/sbin/losetup
func DefaultClient() *Client {
	if client == nil {
		client = NewClient("/usr/sbin/losetup")
	}

	return client
}

// NewClient returns a new instance of the Client
func NewClient(binaryPath string) *Client {
	if binaryPath == "" {
		binaryPath = "/usr/sbin/losetup"
	}

	c := &Client{
		BinaryPath: binaryPath,
	}

	return c
}

// Detach detaches a loop device from a file
func (c *Client) Detach(ctx context.Context, devicePath string) error {
	if _, err := os.Stat(devicePath); err != nil {
		return nil
	}

	_, err := c.losetup(ctx, "--detach", devicePath)
	return err
}

// Attach attaches a loop device to a file
func (c *Client) Attach(ctx context.Context, filePath string) (string, error) {
	existing, err := c.Find(ctx, filePath)
	if err == nil && existing != "" {
		return strings.Split(existing, ":")[0], nil
	}

	output, err := c.losetup(ctx, "--find", "--show", filePath)
	if err != nil {
		return "", err
	}

	return output, nil
}

// Find finds a loop device associated with a file
func (c *Client) Find(ctx context.Context, filePath string) (string, error) {
	output, err := c.losetup(ctx, "--associated", filePath)
	if err != nil {
		return "", err
	}

	output = strings.Split(output, ":")[0]

	return output, nil
}

func (c *Client) losetup(ctx context.Context, args ...string) (string, error) {
	data, err := exec.CommandContext(ctx, c.BinaryPath, args...).CombinedOutput()
	output := string(data)
	if err != nil {
		return "", fmt.Errorf("losetup %s\nerror: %s\n: %w", strings.Join(args, " "), output, err)
	}

	output = strings.TrimSuffix(output, "\n")
	output = strings.TrimSpace(output)

	return output, nil
}
