package containerd

import (
	"github.com/containerd/containerd"
	"github.com/containerd/log"
)

const (
	defaultSnapshotter = "devmapper"
	defaultNamespace   = "fireactions"
)

// Conifg is the configuration for the Client.
type Config struct {
	// Address is the address of the containerd server. If empty, the default
	// address of "/run/containerd/containerd.sock" is used instead.
	Address string
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	c := &Config{
		Address: "/run/containerd/containerd.sock",
	}

	return c
}

// Client is a client that connects to Containerd. It is a wrapper around
// containerd.Client.
type Client struct {
	*containerd.Client

	config *Config
}

// NewClient creates a new Client.
func NewClient(cfg *Config) (*Client, error) {
	client, err := containerd.New(cfg.Address, containerd.WithDefaultNamespace(defaultNamespace))
	if err != nil {
		return nil, err
	}

	log.SetLevel("panic")

	return &Client{Client: client, config: cfg}, nil
}

// Close closes the Client.
func (c *Client) Close() error {
	return c.Client.Close()
}

// ImageService returns an implementation of the Fireactions ImageManager interface. Essentially, it
// is a wrapper around containerd.ImageService.
func (c *Client) ImageService() *imageServiceImpl {
	return &imageServiceImpl{client: c}
}
