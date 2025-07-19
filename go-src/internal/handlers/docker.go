package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"coolify-go/internal/docker"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/gin-gonic/gin"
)

// DockerHandler handles Docker-related HTTP requests
type DockerHandler struct {
	dockerClient *docker.Client
}

// NewDockerHandler creates a new Docker handler
func NewDockerHandler(dockerClient *docker.Client) *DockerHandler {
	return &DockerHandler{
		dockerClient: dockerClient,
	}
}

// RegisterRoutes registers Docker-related routes
func (h *DockerHandler) RegisterRoutes(router *gin.RouterGroup) {
	docker := router.Group("/docker")
	{
		// System information
		docker.GET("/info", h.GetInfo)
		docker.GET("/version", h.GetVersion)
		docker.GET("/ping", h.Ping)

		// Containers
		containers := docker.Group("/containers")
		{
			containers.GET("", h.ListContainers)
			containers.GET("/:id", h.GetContainer)
			containers.POST("", h.CreateContainer)
			containers.POST("/:id/start", h.StartContainer)
			containers.POST("/:id/stop", h.StopContainer)
			containers.DELETE("/:id", h.RemoveContainer)
			containers.GET("/:id/logs", h.GetContainerLogs)
			containers.GET("/:id/stats", h.GetContainerStats)
		}

		// Images
		images := docker.Group("/images")
		{
			images.GET("", h.ListImages)
			images.GET("/:id", h.GetImage)
			images.POST("/pull", h.PullImage)
			images.DELETE("/:id", h.RemoveImage)
			images.POST("/prune", h.PruneImages)
		}

		// Networks
		networks := docker.Group("/networks")
		{
			networks.GET("", h.ListNetworks)
			networks.GET("/:id", h.GetNetwork)
			networks.POST("", h.CreateNetwork)
			networks.DELETE("/:id", h.RemoveNetwork)
		}

		// Swarm
		swarm := docker.Group("/swarm")
		{
			swarm.GET("/info", h.GetSwarmInfo)
			swarm.POST("/init", h.InitSwarm)
			swarm.POST("/join", h.JoinSwarm)
			swarm.POST("/leave", h.LeaveSwarm)

			// Services
			services := swarm.Group("/services")
			{
				services.GET("", h.ListServices)
				services.GET("/:id", h.GetService)
				services.POST("", h.CreateService)
				services.PUT("/:id", h.UpdateService)
				services.DELETE("/:id", h.RemoveService)
			}

			// Nodes
			nodes := swarm.Group("/nodes")
			{
				nodes.GET("", h.ListNodes)
				nodes.GET("/:id", h.GetNode)
				nodes.PUT("/:id", h.UpdateNode)
				nodes.DELETE("/:id", h.RemoveNode)
			}
		}
	}
}

// GetInfo returns Docker system information
func (h *DockerHandler) GetInfo(c *gin.Context) {
	info, err := h.dockerClient.GetInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetVersion returns Docker version information
func (h *DockerHandler) GetVersion(c *gin.Context) {
	version, err := h.dockerClient.GetVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, version)
}

// Ping tests the Docker daemon connection
func (h *DockerHandler) Ping(c *gin.Context) {
	err := h.dockerClient.Ping()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListContainers returns all containers
func (h *DockerHandler) ListContainers(c *gin.Context) {
	all := c.Query("all") == "true"
	containers, err := h.dockerClient.ListContainers(all, filters.NewArgs())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

// GetContainer returns container details
func (h *DockerHandler) GetContainer(c *gin.Context) {
	containerID := c.Param("id")
	container, err := h.dockerClient.GetContainer(containerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, container)
}

// CreateContainer creates a new container
func (h *DockerHandler) CreateContainer(c *gin.Context) {
	var request struct {
		Image       string            `json:"image" binding:"required"`
		Name        string            `json:"name"`
		Environment map[string]string `json:"environment"`
		Ports       []string          `json:"ports"`
		Volumes     map[string]string `json:"volumes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create container config
	config := &container.Config{
		Image: request.Image,
		Env:   convertEnvVars(request.Environment),
	}

	// Create host config
	hostConfig := &container.HostConfig{
		Binds: convertVolumeBinds(request.Volumes),
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	containerID, err := h.dockerClient.CreateContainer(config, hostConfig, request.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": containerID})
}

// StartContainer starts a container
func (h *DockerHandler) StartContainer(c *gin.Context) {
	containerID := c.Param("id")
	err := h.dockerClient.StartContainer(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

// StopContainer stops a container
func (h *DockerHandler) StopContainer(c *gin.Context) {
	containerID := c.Param("id")
	timeout := c.Query("timeout")

	var timeoutDuration *time.Duration
	if timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			duration := time.Duration(t) * time.Second
			timeoutDuration = &duration
		}
	}

	err := h.dockerClient.StopContainer(containerID, timeoutDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

// RemoveContainer removes a container
func (h *DockerHandler) RemoveContainer(c *gin.Context) {
	containerID := c.Param("id")
	force := c.Query("force") == "true"

	err := h.dockerClient.RemoveContainer(containerID, force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// GetContainerLogs returns container logs
func (h *DockerHandler) GetContainerLogs(c *gin.Context) {
	containerID := c.Param("id")
	tail := c.DefaultQuery("tail", "100")
	follow := c.Query("follow") == "true"

	logs, err := h.dockerClient.GetContainerLogs(containerID, tail, follow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer logs.Close()

	c.DataFromReader(http.StatusOK, -1, "text/plain", logs, nil)
}

// GetContainerStats returns container statistics
func (h *DockerHandler) GetContainerStats(c *gin.Context) {
	containerID := c.Param("id")
	stream := c.Query("stream") == "true"

	stats, err := h.dockerClient.GetContainerStats(containerID, stream)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stats.Close()

	c.DataFromReader(http.StatusOK, -1, "application/json", stats, nil)
}

// ListImages returns all images
func (h *DockerHandler) ListImages(c *gin.Context) {
	images, err := h.dockerClient.ListImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

// GetImage returns image details
func (h *DockerHandler) GetImage(c *gin.Context) {
	imageID := c.Param("id")
	image, err := h.dockerClient.GetImage(imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, image)
}

// PullImage pulls an image from registry
func (h *DockerHandler) PullImage(c *gin.Context) {
	var request struct {
		Image string `json:"image" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.dockerClient.PullImage(request.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "pulled"})
}

// RemoveImage removes an image
func (h *DockerHandler) RemoveImage(c *gin.Context) {
	imageID := c.Param("id")
	force := c.Query("force") == "true"

	err := h.dockerClient.RemoveImage(imageID, force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// PruneImages removes unused images
func (h *DockerHandler) PruneImages(c *gin.Context) {
	report, err := h.dockerClient.PruneImages(filters.NewArgs())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// ListNetworks returns all networks
func (h *DockerHandler) ListNetworks(c *gin.Context) {
	networks, err := h.dockerClient.ListNetworks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, networks)
}

// GetNetwork returns network details
func (h *DockerHandler) GetNetwork(c *gin.Context) {
	networkID := c.Param("id")
	network, err := h.dockerClient.GetNetwork(networkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, network)
}

// CreateNetwork creates a new network
func (h *DockerHandler) CreateNetwork(c *gin.Context) {
	var request struct {
		Name   string `json:"name" binding:"required"`
		Driver string `json:"driver"`
		Subnet string `json:"subnet"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driver := request.Driver
	if driver == "" {
		driver = "bridge"
	}

	var resp types.NetworkCreateResponse
	var err error

	if driver == "overlay" {
		resp, err = h.dockerClient.CreateOverlayNetwork(request.Name, types.NetworkCreate{})
	} else {
		resp, err = h.dockerClient.CreateBridgeNetwork(request.Name, request.Subnet)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// RemoveNetwork removes a network
func (h *DockerHandler) RemoveNetwork(c *gin.Context) {
	networkID := c.Param("id")
	err := h.dockerClient.RemoveNetwork(networkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// GetSwarmInfo returns swarm cluster information
func (h *DockerHandler) GetSwarmInfo(c *gin.Context) {
	info, err := h.dockerClient.GetSwarmInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// InitSwarm initializes a new swarm cluster
func (h *DockerHandler) InitSwarm(c *gin.Context) {
	var request swarm.InitRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.dockerClient.InitSwarm(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// JoinSwarm joins a node to swarm cluster
func (h *DockerHandler) JoinSwarm(c *gin.Context) {
	var request swarm.JoinRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.dockerClient.JoinSwarm(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "joined"})
}

// LeaveSwarm leaves the swarm cluster
func (h *DockerHandler) LeaveSwarm(c *gin.Context) {
	force := c.Query("force") == "true"

	err := h.dockerClient.LeaveSwarm(force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "left"})
}

// ListServices returns all swarm services
func (h *DockerHandler) ListServices(c *gin.Context) {
	services, err := h.dockerClient.ListServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services)
}

// GetService returns service details
func (h *DockerHandler) GetService(c *gin.Context) {
	serviceID := c.Param("id")
	service, err := h.dockerClient.GetService(serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, service)
}

// CreateService creates a new swarm service
func (h *DockerHandler) CreateService(c *gin.Context) {
	var spec swarm.ServiceSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.dockerClient.CreateService(spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateService updates a swarm service
func (h *DockerHandler) UpdateService(c *gin.Context) {
	serviceID := c.Param("id")

	var request struct {
		Version swarm.Version     `json:"version"`
		Spec    swarm.ServiceSpec `json:"spec"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.dockerClient.UpdateService(serviceID, request.Version, request.Spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// RemoveService removes a swarm service
func (h *DockerHandler) RemoveService(c *gin.Context) {
	serviceID := c.Param("id")

	err := h.dockerClient.RemoveService(serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// ListNodes returns all swarm nodes
func (h *DockerHandler) ListNodes(c *gin.Context) {
	nodes, err := h.dockerClient.ListNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// GetNode returns node details
func (h *DockerHandler) GetNode(c *gin.Context) {
	nodeID := c.Param("id")
	node, err := h.dockerClient.GetNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, node)
}

// UpdateNode updates a swarm node
func (h *DockerHandler) UpdateNode(c *gin.Context) {
	nodeID := c.Param("id")

	var request struct {
		Version swarm.Version  `json:"version"`
		Spec    swarm.NodeSpec `json:"spec"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.dockerClient.UpdateNode(nodeID, request.Version, request.Spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// RemoveNode removes a swarm node
func (h *DockerHandler) RemoveNode(c *gin.Context) {
	nodeID := c.Param("id")
	force := c.Query("force") == "true"

	err := h.dockerClient.RemoveNode(nodeID, force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// Helper functions
func convertEnvVars(envVars map[string]string) []string {
	var env []string
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

func convertVolumeBinds(volumes map[string]string) []string {
	var binds []string
	for hostPath, containerPath := range volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}
	return binds
}
