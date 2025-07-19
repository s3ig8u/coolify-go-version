package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListNetworks(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	networks, err := client.ListNetworks()
	assert.NoError(t, err)
	assert.NotNil(t, networks)
}

func TestClient_GetNetwork(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// First, list networks to get an existing network ID
	networks, err := client.ListNetworks()
	if err != nil {
		t.Skip("No networks available for testing")
	}

	if len(networks) == 0 {
		t.Skip("No networks available for testing")
	}

	networkID := networks[0].ID
	network, err := client.GetNetwork(networkID)
	assert.NoError(t, err)
	assert.NotNil(t, network)
	assert.Equal(t, networkID, network.ID)
}

func TestClient_CreateNetwork(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a test network
	networkName := "test-network"
	driver := "bridge"
	options := types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: "172.20.0.0/16",
				},
			},
		},
	}

	resp, err := client.CreateNetwork(networkName, driver, options)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	// Verify the network was created
	networks, err := client.ListNetworks()
	assert.NoError(t, err)

	found := false
	for _, network := range networks {
		if network.Name == networkName {
			found = true
			break
		}
	}
	assert.True(t, found, "Created network should be in the list")

	// Clean up
	defer func() {
		client.RemoveNetwork(resp.ID)
	}()
}

func TestClient_RemoveNetwork(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// First, create a network to remove
	networkName := "test-remove-network"
	driver := "bridge"
	options := types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: "172.21.0.0/16",
				},
			},
		},
	}

	resp, err := client.CreateNetwork(networkName, driver, options)
	require.NoError(t, err)

	// Verify the network exists
	networks, err := client.ListNetworks()
	require.NoError(t, err)

	found := false
	for _, network := range networks {
		if network.Name == networkName {
			found = true
			break
		}
	}
	require.True(t, found, "Network should exist before removal")

	// Remove the network
	err = client.RemoveNetwork(resp.ID)
	assert.NoError(t, err)

	// Verify the network was removed
	networks, err = client.ListNetworks()
	assert.NoError(t, err)

	found = false
	for _, network := range networks {
		if network.Name == networkName {
			found = true
			break
		}
	}
	assert.False(t, found, "Network should not exist after removal")
}

func TestClient_ConnectContainerToNetwork(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a test network
	networkName := "test-connect-network"
	driver := "bridge"
	options := types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: "172.22.0.0/16",
				},
			},
		},
	}

	resp, err := client.CreateNetwork(networkName, driver, options)
	require.NoError(t, err)
	defer func() {
		client.RemoveNetwork(resp.ID)
	}()

	// Create a test container
	containerConfig := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "30"},
	}

	containerID, err := client.CreateContainer(containerConfig, nil, "test-connect-container")
	require.NoError(t, err)
	defer func() {
		client.RemoveContainer(containerID, true)
	}()

	// Connect container to network
	err = client.ConnectContainerToNetwork(containerID, resp.ID, nil)
	assert.NoError(t, err)

	// Verify the connection
	network, err := client.GetNetwork(resp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, network)

	// Check if container is connected
	connected := false
	for containerName := range network.Containers {
		if containerName == "test-connect-container" {
			connected = true
			break
		}
	}
	assert.True(t, connected, "Container should be connected to network")
}

func TestClient_DisconnectContainerFromNetwork(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a test network
	networkName := "test-disconnect-network"
	driver := "bridge"
	options := types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: "172.23.0.0/16",
				},
			},
		},
	}

	resp, err := client.CreateNetwork(networkName, driver, options)
	require.NoError(t, err)
	defer func() {
		client.RemoveNetwork(resp.ID)
	}()

	// Create a test container
	containerConfig := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "30"},
	}

	containerID, err := client.CreateContainer(containerConfig, nil, "test-disconnect-container")
	require.NoError(t, err)
	defer func() {
		client.RemoveContainer(containerID, true)
	}()

	// Connect container to network
	err = client.ConnectContainerToNetwork(containerID, resp.ID, nil)
	require.NoError(t, err)

	// Disconnect container from network
	err = client.DisconnectContainerFromNetwork(containerID, resp.ID, false)
	assert.NoError(t, err)

	// Verify the disconnection
	network, err := client.GetNetwork(resp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, network)

	// Check if container is disconnected
	connected := false
	for containerName := range network.Containers {
		if containerName == "test-disconnect-container" {
			connected = true
			break
		}
	}
	assert.False(t, connected, "Container should be disconnected from network")
}

func TestClient_CreateBridgeNetwork(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a bridge network
	networkName := "test-bridge-network"
	subnet := "172.24.0.0/16"

	resp, err := client.CreateBridgeNetwork(networkName, subnet)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	// Verify the network was created
	networks, err := client.ListNetworks()
	assert.NoError(t, err)

	found := false
	for _, network := range networks {
		if network.Name == networkName {
			found = true
			break
		}
	}
	assert.True(t, found, "Created bridge network should be in the list")

	// Clean up
	defer func() {
		client.RemoveNetwork(resp.ID)
	}()
}
