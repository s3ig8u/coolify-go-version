package docker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// Test with default environment (requires Docker to be running)
	client, err := NewClient()

	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.GetClient())
}

func TestDockerClient_Connect(t *testing.T) {
	// Test with mock Docker daemon (requires Docker to be running)
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	err = client.Ping()
	assert.NoError(t, err)
}

func TestDockerClient_Close(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	err = client.Close()
	assert.NoError(t, err)
}

func TestDockerClient_GetInfo(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	info, err := client.GetInfo()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	assert.NotNil(t, info)
	assert.NotEmpty(t, info.Name)
	assert.NotEmpty(t, info.OperatingSystem)
}

func TestDockerClient_GetVersion(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	version, err := client.GetVersion()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	assert.NotNil(t, version)
	assert.NotEmpty(t, version.Version)
	assert.NotEmpty(t, version.APIVersion)
}

func TestDockerClient_Ping(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	err = client.Ping()
	assert.NoError(t, err)
}

func TestDockerClient_WithTimeout(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test that timeout context works
	_, err = client.GetClient().Ping(ctx)
	assert.NoError(t, err)
}

func TestDockerClient_GetClientAndContext(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Test GetClient
	dockerClient := client.GetClient()
	assert.NotNil(t, dockerClient)

	// Test GetContext
	ctx := client.GetContext()
	assert.NotNil(t, ctx)
}
