package docker

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListContainers(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	containers, err := client.ListContainers(true, filters.Args{})
	assert.NoError(t, err)
	assert.NotNil(t, containers)
}

func TestClient_GetContainer(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// First, list containers to get an existing container ID
	containers, err := client.ListContainers(true, filters.Args{})
	if err != nil {
		t.Skip("No containers available for testing")
	}

	if len(containers) == 0 {
		t.Skip("No containers available for testing")
	}

	containerID := containers[0].ID
	container, err := client.GetContainer(containerID)
	assert.NoError(t, err)
	assert.NotNil(t, container)
	assert.Equal(t, containerID, container.ID)
}

func TestClient_CreateContainer(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a simple test container
	config := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "hello world"},
	}

	containerID, err := client.CreateContainer(config, nil, "test-container")
	assert.NoError(t, err)
	assert.NotEmpty(t, containerID)

	// Clean up
	defer func() {
		client.RemoveContainer(containerID, true)
	}()
}

func TestClient_StartContainer(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a simple test container
	config := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "10"},
	}

	containerID, err := client.CreateContainer(config, nil, "test-start-container")
	require.NoError(t, err)
	defer func() {
		client.RemoveContainer(containerID, true)
	}()

	// Start the container
	err = client.StartContainer(containerID)
	assert.NoError(t, err)

	// Verify container is running
	container, err := client.GetContainer(containerID)
	assert.NoError(t, err)
	assert.Equal(t, "running", container.State.Status)
}

func TestClient_StopContainer(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create and start a test container
	config := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "30"},
	}

	containerID, err := client.CreateContainer(config, nil, "test-stop-container")
	require.NoError(t, err)
	defer func() {
		client.RemoveContainer(containerID, true)
	}()

	err = client.StartContainer(containerID)
	require.NoError(t, err)

	// Stop the container
	timeout := 10 * time.Second
	err = client.StopContainer(containerID, &timeout)
	assert.NoError(t, err)

	// Verify container is stopped
	container, err := client.GetContainer(containerID)
	assert.NoError(t, err)
	assert.Equal(t, "exited", container.State.Status)
}

func TestClient_RemoveContainer(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a test container
	config := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "hello world"},
	}

	containerID, err := client.CreateContainer(config, nil, "test-remove-container")
	require.NoError(t, err)

	// Remove the container
	err = client.RemoveContainer(containerID, false)
	assert.NoError(t, err)

	// Verify container is removed
	_, err = client.GetContainer(containerID)
	assert.Error(t, err) // Should fail because container no longer exists
}

func TestClient_GetContainerLogs(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create and start a test container that produces output
	config := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "test log output"},
	}

	containerID, err := client.CreateContainer(config, nil, "test-logs-container")
	require.NoError(t, err)
	defer func() {
		client.RemoveContainer(containerID, true)
	}()

	err = client.StartContainer(containerID)
	require.NoError(t, err)

	// Wait a moment for the container to finish
	time.Sleep(2 * time.Second)

	// Get container logs
	logs, err := client.GetContainerLogs(containerID, "100", false)
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	defer logs.Close()
}

func TestClient_GetContainerStats(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create and start a test container
	config := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "10"},
	}

	containerID, err := client.CreateContainer(config, nil, "test-stats-container")
	require.NoError(t, err)
	defer func() {
		client.RemoveContainer(containerID, true)
	}()

	err = client.StartContainer(containerID)
	require.NoError(t, err)

	// Get container stats
	stats, err := client.GetContainerStats(containerID, false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	defer stats.Close()
}

func TestClient_ExecuteCommand(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create and start a test container
	config := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "30"},
	}

	containerID, err := client.CreateContainer(config, nil, "test-exec-container")
	require.NoError(t, err)
	defer func() {
		client.RemoveContainer(containerID, true)
	}()

	err = client.StartContainer(containerID)
	require.NoError(t, err)

	// Execute a command in the container
	cmd := []string{"echo", "hello from exec"}
	resp, err := client.ExecuteCommand(containerID, cmd)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}
