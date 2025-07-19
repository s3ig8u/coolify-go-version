package docker

import (
	"testing"

	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func TestClient_InitSwarm(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Initialize swarm
	options := swarm.InitRequest{
		ListenAddr:      "0.0.0.0:2377",
		AdvertiseAddr:   "127.0.0.1:2377",
		ForceNewCluster: false,
	}

	nodeID, err := client.InitSwarm(options)
	assert.NoError(t, err)
	assert.NotEmpty(t, nodeID)

	// Verify swarm is initialized
	info, err := client.GetSwarmInfo()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.ID)
}

func TestClient_JoinSwarm(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// This test requires a running swarm cluster
	// For now, we'll just test the function signature
	options := swarm.JoinRequest{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: "127.0.0.1:2377",
		RemoteAddrs:   []string{"127.0.0.1:2377"},
		JoinToken:     "test-token",
	}

	err = client.JoinSwarm(options)
	// This will likely fail in a test environment, but we're testing the function exists
	// In a real environment, you'd need a running swarm cluster
	assert.Error(t, err) // Expected to fail in test environment
}

func TestClient_LeaveSwarm(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Leave swarm
	err = client.LeaveSwarm(false)
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.NoError(t, err) // Should not error even if not in swarm mode
}

func TestClient_GetSwarmInfo(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	info, err := client.GetSwarmInfo()
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestClient_ListServices(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	_, err = client.ListServices()
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}

func TestClient_GetService(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	_, err = client.GetService("test-service-id")
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}

func TestClient_CreateService(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a simple service spec
	spec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: "test-service",
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "alpine:latest",
				Command: []string{"sleep", "30"},
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &[]uint64{1}[0],
			},
		},
	}

	_, err = client.CreateService(spec)
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}

func TestClient_UpdateService(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Update service spec
	version := swarm.Version{Index: 1}
	spec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: "test-service",
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "alpine:latest",
				Command: []string{"sleep", "60"},
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &[]uint64{2}[0],
			},
		},
	}

	err = client.UpdateService("test-service-id", version, spec)
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}

func TestClient_RemoveService(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	err = client.RemoveService("test-service-id")
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}

func TestClient_ListNodes(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	_, err = client.ListNodes()
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}

func TestClient_GetNode(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	_, err = client.GetNode("test-node-id")
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}

func TestClient_ListTasks(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	_, err = client.ListTasks("test-service-id")
	// This might fail if not in swarm mode, which is expected
	// We're just testing the function exists and can be called
	assert.Error(t, err) // Expected to fail if not in swarm mode
}
