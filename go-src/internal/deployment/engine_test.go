package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentEngine_ConvertEnvVars(t *testing.T) {
	engine := &Engine{}

	envVars := map[string]string{
		"NODE_ENV": "production",
		"PORT":     "8080",
	}

	result := engine.convertEnvVars(envVars)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "NODE_ENV=production")
	assert.Contains(t, result, "PORT=8080")
}

func TestDeploymentEngine_ConvertPorts(t *testing.T) {
	engine := &Engine{}

	ports := []string{"8080:8080", "3000:3000"}

	result := engine.convertPorts(ports)
	assert.Len(t, result, 2)
	assert.Equal(t, uint32(8080), result[0].PublishedPort)
	assert.Equal(t, uint32(8080), result[0].TargetPort)
	assert.Equal(t, uint32(3000), result[1].PublishedPort)
	assert.Equal(t, uint32(3000), result[1].TargetPort)
}

func TestDeploymentEngine_ConvertVolumeBinds(t *testing.T) {
	engine := &Engine{}

	volumes := map[string]string{
		"/tmp": "/data",
		"/var": "/container",
	}

	result := engine.convertVolumeBinds(volumes)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "/tmp:/data")
	assert.Contains(t, result, "/var:/container")
}

func TestDeploymentEngine_ConfigStructure(t *testing.T) {
	// Test configuration structure
	config := &Config{
		Repository:      "https://github.com/test/repo.git",
		Branch:          "main",
		Environment:     "production",
		BuildArgs:       map[string]*string{"NODE_ENV": stringPtr("production")},
		Ports:           []string{"8080:8080"},
		Volumes:         map[string]string{"/tmp": "/data"},
		EnvironmentVars: map[string]string{"NODE_ENV": "production"},
		HealthCheck:     "http://localhost:8080/health",
		Replicas:        1,
		Network:         "default",
	}

	assert.Len(t, config.BuildArgs, 1)
	assert.Equal(t, "production", *config.BuildArgs["NODE_ENV"])
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
