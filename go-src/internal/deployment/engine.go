package deployment

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"coolify-go/internal/docker"
	"coolify-go/internal/ssh"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

// Engine represents the deployment engine
type Engine struct {
	dockerClient *docker.Client
	sshClient    *ssh.Client
	workDir      string
}

// Config represents deployment configuration
type Config struct {
	Repository      string
	Branch          string
	Commit          string
	Environment     string
	BuildArgs       map[string]*string
	Ports           []string
	Volumes         map[string]string
	EnvironmentVars map[string]string
	HealthCheck     string
	Replicas        int
	Network         string
}

// DeploymentResult represents the result of a deployment
type DeploymentResult struct {
	Success     bool
	ContainerID string
	ServiceID   string
	Logs        string
	Error       error
	DeployedAt  time.Time
}

// NewEngine creates a new deployment engine
func NewEngine(dockerClient *docker.Client, sshClient *ssh.Client, workDir string) *Engine {
	return &Engine{
		dockerClient: dockerClient,
		sshClient:    sshClient,
		workDir:      workDir,
	}
}

// DeployApplication deploys an application using the specified configuration
func (e *Engine) DeployApplication(config *Config) (*DeploymentResult, error) {
	result := &DeploymentResult{
		DeployedAt: time.Now(),
	}

	logrus.Infof("🚀 Starting deployment for %s", config.Repository)

	// Step 1: Clone or update repository
	if err := e.cloneRepository(config); err != nil {
		result.Error = fmt.Errorf("failed to clone repository: %w", err)
		return result, result.Error
	}

	// Step 2: Build Docker image
	imageName, err := e.buildImage(config)
	if err != nil {
		result.Error = fmt.Errorf("failed to build image: %w", err)
		return result, result.Error
	}

	// Step 3: Deploy container/service
	if err := e.deployContainer(config, imageName, result); err != nil {
		result.Error = fmt.Errorf("failed to deploy container: %w", err)
		return result, result.Error
	}

	result.Success = true
	logrus.Infof("✅ Deployment completed successfully")
	return result, nil
}

// cloneRepository clones or updates the Git repository
func (e *Engine) cloneRepository(config *Config) error {
	repoPath := filepath.Join(e.workDir, "repos", strings.ReplaceAll(config.Repository, "/", "_"))

	// Check if repository already exists
	if _, err := os.Stat(repoPath); err == nil {
		logrus.Infof("📁 Repository exists, updating...")

		// Pull latest changes
		cmd := exec.Command("git", "pull", "origin", config.Branch)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to pull repository: %s, %w", string(output), err)
		}
	} else {
		logrus.Infof("📁 Cloning repository...")

		// Create directory
		if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Clone repository
		cmd := exec.Command("git", "clone", "-b", config.Branch, config.Repository, repoPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to clone repository: %s, %w", string(output), err)
		}
	}

	// Checkout specific commit if provided
	if config.Commit != "" {
		cmd := exec.Command("git", "checkout", config.Commit)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout commit: %s, %w", string(output), err)
		}
	}

	logrus.Infof("✅ Repository ready: %s", repoPath)
	return nil
}

// buildImage builds a Docker image from the repository
func (e *Engine) buildImage(config *Config) (string, error) {
	repoPath := filepath.Join(e.workDir, "repos", strings.ReplaceAll(config.Repository, "/", "_"))
	imageName := fmt.Sprintf("coolify-app:%s", strings.ReplaceAll(config.Repository, "/", "-"))

	logrus.Infof("🔨 Building Docker image: %s", imageName)

	// Check for Dockerfile
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err != nil {
		return "", fmt.Errorf("Dockerfile not found in repository")
	}

	// Build context
	buildContext, err := os.Open(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open build context: %w", err)
	}
	defer buildContext.Close()

	// Build options
	buildOptions := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageName},
		BuildArgs:  config.BuildArgs,
		Remove:     true,
	}

	// Build image
	resp, err := e.dockerClient.BuildImage(buildContext, buildOptions)
	if err != nil {
		return "", err
	}

	// Read build logs
	if resp.Body != nil {
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)
	}

	logrus.Infof("✅ Image built successfully: %s", imageName)
	return imageName, nil
}

// deployContainer deploys the container or service
func (e *Engine) deployContainer(config *Config, imageName string, result *DeploymentResult) error {
	appName := strings.ReplaceAll(config.Repository, "/", "-")

	// Check if we're in Swarm mode
	swarmInfo, err := e.dockerClient.GetSwarmInfo()
	if err == nil && swarmInfo != nil {
		// Deploy as Swarm service
		return e.deploySwarmService(config, imageName, appName, result)
	} else {
		// Deploy as standalone container
		return e.deployStandaloneContainer(config, imageName, appName, result)
	}
}

// deploySwarmService deploys as a Swarm service
func (e *Engine) deploySwarmService(config *Config, imageName, appName string, result *DeploymentResult) error {
	logrus.Infof("🐳 Deploying as Swarm service: %s", appName)

	// Convert replicas to uint64
	replicas := uint64(config.Replicas)

	// Create service spec
	spec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: appName,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: imageName,
				Env:   e.convertEnvVars(config.EnvironmentVars),
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		UpdateConfig: &swarm.UpdateConfig{
			Order: "start-first",
		},
		RollbackConfig: &swarm.UpdateConfig{
			Order: "start-first",
		},
	}

	// Add port mappings
	if len(config.Ports) > 0 {
		spec.EndpointSpec = &swarm.EndpointSpec{
			Ports: e.convertPorts(config.Ports),
		}
	}

	// Create service
	resp, err := e.dockerClient.CreateService(spec)
	if err != nil {
		return err
	}

	result.ServiceID = resp.ID
	logrus.Infof("✅ Swarm service created: %s", resp.ID)
	return nil
}

// deployStandaloneContainer deploys as a standalone container
func (e *Engine) deployStandaloneContainer(config *Config, imageName, appName string, result *DeploymentResult) error {
	logrus.Infof("🐳 Deploying as standalone container: %s", appName)

	// Stop existing container if it exists
	containers, err := e.dockerClient.ListContainers(true, filters.NewArgs(filters.Arg("name", appName)))
	if err == nil && len(containers) > 0 {
		for _, container := range containers {
			e.dockerClient.StopContainer(container.ID, nil)
			e.dockerClient.RemoveContainer(container.ID, true)
		}
	}

	// Create container config
	containerConfig := &container.Config{
		Image: imageName,
		Env:   e.convertEnvVars(config.EnvironmentVars),
	}

	// Add port mappings
	if len(config.Ports) > 0 {
		containerConfig.ExposedPorts = e.convertExposedPorts(config.Ports)
	}

	// Create host config
	hostConfig := &container.HostConfig{
		PortBindings: e.convertPortBindings(config.Ports),
		Binds:        e.convertVolumeBinds(config.Volumes),
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Create container
	containerID, err := e.dockerClient.CreateContainer(containerConfig, hostConfig, appName)
	if err != nil {
		return err
	}

	// Start container
	if err := e.dockerClient.StartContainer(containerID); err != nil {
		return err
	}

	result.ContainerID = containerID
	logrus.Infof("✅ Container created and started: %s", containerID)
	return nil
}

// convertEnvVars converts environment variables map to slice
func (e *Engine) convertEnvVars(envVars map[string]string) []string {
	var env []string
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// convertPorts converts port strings to swarm port configs
func (e *Engine) convertPorts(ports []string) []swarm.PortConfig {
	var portConfigs []swarm.PortConfig
	for _, port := range ports {
		parts := strings.Split(port, ":")
		if len(parts) == 2 {
			publishedPort, _ := strconv.ParseUint(parts[0], 10, 32)
			targetPort, _ := strconv.ParseUint(parts[1], 10, 32)
			portConfigs = append(portConfigs, swarm.PortConfig{
				PublishedPort: uint32(publishedPort),
				TargetPort:    uint32(targetPort),
				Protocol:      "tcp",
			})
		}
	}
	return portConfigs
}

// convertExposedPorts converts port strings to exposed ports
func (e *Engine) convertExposedPorts(ports []string) nat.PortSet {
	portSet := make(nat.PortSet)
	for _, port := range ports {
		parts := strings.Split(port, ":")
		if len(parts) == 2 {
			portSet[nat.Port(parts[1]+"/tcp")] = struct{}{}
		}
	}
	return portSet
}

// convertPortBindings converts port strings to port bindings
func (e *Engine) convertPortBindings(ports []string) nat.PortMap {
	portMap := make(nat.PortMap)
	for _, port := range ports {
		parts := strings.Split(port, ":")
		if len(parts) == 2 {
			portMap[nat.Port(parts[1]+"/tcp")] = []nat.PortBinding{
				{HostPort: parts[0]},
			}
		}
	}
	return portMap
}

// convertVolumeBinds converts volume map to binds slice
func (e *Engine) convertVolumeBinds(volumes map[string]string) []string {
	var binds []string
	for hostPath, containerPath := range volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}
	return binds
}
