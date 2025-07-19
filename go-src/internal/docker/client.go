package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

// Client wraps the Docker client with additional functionality
type Client struct {
	client *client.Client
	ctx    context.Context
}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	ctx := context.Background()

	// Create Docker client with default settings
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithTimeout(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Test the connection
	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}

	logrus.Info("✅ Docker client connected successfully")

	return &Client{
		client: cli,
		ctx:    ctx,
	}, nil
}

// Close closes the Docker client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Docker client
func (c *Client) GetClient() *client.Client {
	return c.client
}

// GetContext returns the client context
func (c *Client) GetContext() context.Context {
	return c.ctx
}

// Ping tests the connection to the Docker daemon
func (c *Client) Ping() error {
	_, err := c.client.Ping(c.ctx)
	return err
}

// GetInfo returns Docker system information
func (c *Client) GetInfo() (system.Info, error) {
	return c.client.Info(c.ctx)
}

// GetVersion returns Docker version information
func (c *Client) GetVersion() (types.Version, error) {
	return c.client.ServerVersion(c.ctx)
}
