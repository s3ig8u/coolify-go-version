package docker

import (
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/sirupsen/logrus"
)

// ContainerInfo represents container information
type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	State   string            `json:"state"`
	Created int64             `json:"created"`
	Ports   []types.Port      `json:"ports"`
	Labels  map[string]string `json:"labels"`
	Network string            `json:"network"`
}

// ListContainers returns all containers with optional filters
func (c *Client) ListContainers(all bool, filters filters.Args) ([]ContainerInfo, error) {
	containers, err := c.client.ContainerList(c.ctx, container.ListOptions{
		All:     all,
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var containerInfos []ContainerInfo
	for _, container := range containers {
		info := ContainerInfo{
			ID:      container.ID,
			Name:    container.Names[0], // First name (without leading slash)
			Image:   container.Image,
			Status:  container.Status,
			State:   container.State,
			Created: container.Created,
			Ports:   container.Ports,
			Labels:  container.Labels,
		}

		// Get network info
		if len(container.NetworkSettings.Networks) > 0 {
			for networkName := range container.NetworkSettings.Networks {
				info.Network = networkName
				break
			}
		}

		containerInfos = append(containerInfos, info)
	}

	return containerInfos, nil
}

// GetContainer returns detailed information about a specific container
func (c *Client) GetContainer(containerID string) (*types.ContainerJSON, error) {
	container, err := c.client.ContainerInspect(c.ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}
	return &container, nil
}

// CreateContainer creates a new container
func (c *Client) CreateContainer(config *container.Config, hostConfig *container.HostConfig, name string) (string, error) {
	resp, err := c.client.ContainerCreate(c.ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	logrus.Infof("✅ Container created: %s", resp.ID)
	return resp.ID, nil
}

// StartContainer starts a container
func (c *Client) StartContainer(containerID string) error {
	err := c.client.ContainerStart(c.ctx, containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container %s: %w", containerID, err)
	}

	logrus.Infof("✅ Container started: %s", containerID)
	return nil
}

// StopContainer stops a container with optional timeout
func (c *Client) StopContainer(containerID string, timeout *time.Duration) error {
	var timeoutSeconds *int
	if timeout != nil {
		seconds := int(timeout.Seconds())
		timeoutSeconds = &seconds
	}

	err := c.client.ContainerStop(c.ctx, containerID, container.StopOptions{
		Timeout: timeoutSeconds,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	logrus.Infof("✅ Container stopped: %s", containerID)
	return nil
}

// RemoveContainer removes a container
func (c *Client) RemoveContainer(containerID string, force bool) error {
	err := c.client.ContainerRemove(c.ctx, containerID, container.RemoveOptions{
		Force: force,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}

	logrus.Infof("✅ Container removed: %s", containerID)
	return nil
}

// GetContainerLogs returns container logs
func (c *Client) GetContainerLogs(containerID string, tail string, follow bool) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
	}

	logs, err := c.client.ContainerLogs(c.ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for container %s: %w", containerID, err)
	}

	return logs, nil
}

// ExecuteCommand executes a command in a running container
func (c *Client) ExecuteCommand(containerID string, cmd []string) (types.IDResponse, error) {
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	resp, err := c.client.ContainerExecCreate(c.ctx, containerID, execConfig)
	if err != nil {
		return types.IDResponse{}, fmt.Errorf("failed to create exec in container %s: %w", containerID, err)
	}

	return resp, nil
}

// GetContainerStats returns real-time container statistics
func (c *Client) GetContainerStats(containerID string, stream bool) (io.ReadCloser, error) {
	stats, err := c.client.ContainerStats(c.ctx, containerID, stream)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for container %s: %w", containerID, err)
	}

	return stats.Body, nil
}
