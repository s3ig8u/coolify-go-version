package docker

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/sirupsen/logrus"
)

// SwarmInfo represents swarm cluster information
type SwarmInfo struct {
	ID                     string        `json:"id"`
	Version                swarm.Version `json:"version"`
	CreatedAt              string        `json:"created_at"`
	UpdatedAt              string        `json:"updated_at"`
	Spec                   swarm.Spec    `json:"spec"`
	TLSInfo                swarm.TLSInfo `json:"tls_info"`
	RootRotationInProgress bool          `json:"root_rotation_in_progress"`
	DataPathPort           uint32        `json:"data_path_port"`
	DefaultAddrPool        []string      `json:"default_addr_pool"`
	SubnetSize             uint32        `json:"subnet_size"`
}

// ServiceInfo represents swarm service information
type ServiceInfo struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Image        string              `json:"image"`
	Replicas     uint64              `json:"replicas"`
	Ports        []swarm.PortConfig  `json:"ports"`
	Labels       map[string]string   `json:"labels"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	Spec         swarm.ServiceSpec   `json:"spec"`
	Endpoint     swarm.Endpoint      `json:"endpoint"`
	UpdateStatus *swarm.UpdateStatus `json:"update_status"`
}

// GetSwarmInfo returns swarm cluster information
func (c *Client) GetSwarmInfo() (*SwarmInfo, error) {
	info, err := c.client.SwarmInspect(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect swarm: %w", err)
	}

	swarmInfo := &SwarmInfo{
		ID:                     info.ID,
		Version:                info.Version,
		CreatedAt:              info.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              info.UpdatedAt.Format(time.RFC3339),
		Spec:                   info.Spec,
		TLSInfo:                info.TLSInfo,
		RootRotationInProgress: info.RootRotationInProgress,
		DataPathPort:           info.DataPathPort,
		DefaultAddrPool:        info.DefaultAddrPool,
		SubnetSize:             info.SubnetSize,
	}

	return swarmInfo, nil
}

// InitSwarm initializes a new swarm cluster
func (c *Client) InitSwarm(req swarm.InitRequest) (string, error) {
	resp, err := c.client.SwarmInit(c.ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to initialize swarm: %w", err)
	}

	logrus.Infof("✅ Swarm initialized: %s", resp)
	return resp, nil
}

// JoinSwarm joins a node to an existing swarm cluster
func (c *Client) JoinSwarm(req swarm.JoinRequest) error {
	err := c.client.SwarmJoin(c.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to join swarm: %w", err)
	}

	logrus.Info("✅ Node joined swarm successfully")
	return nil
}

// LeaveSwarm leaves the swarm cluster
func (c *Client) LeaveSwarm(force bool) error {
	err := c.client.SwarmLeave(c.ctx, force)
	if err != nil {
		return fmt.Errorf("failed to leave swarm: %w", err)
	}

	logrus.Info("✅ Node left swarm successfully")
	return nil
}

// ListServices returns all swarm services
func (c *Client) ListServices() ([]ServiceInfo, error) {
	services, err := c.client.ServiceList(c.ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var serviceInfos []ServiceInfo
	for _, service := range services {
		info := ServiceInfo{
			ID:           service.ID,
			Name:         service.Spec.Name,
			Image:        service.Spec.TaskTemplate.ContainerSpec.Image,
			Replicas:     *service.Spec.Mode.Replicated.Replicas,
			Ports:        service.Spec.EndpointSpec.Ports,
			Labels:       service.Spec.Labels,
			CreatedAt:    service.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    service.UpdatedAt.Format(time.RFC3339),
			Spec:         service.Spec,
			Endpoint:     service.Endpoint,
			UpdateStatus: service.UpdateStatus,
		}
		serviceInfos = append(serviceInfos, info)
	}

	return serviceInfos, nil
}

// GetService returns detailed information about a specific service
func (c *Client) GetService(serviceID string) (*swarm.Service, error) {
	service, _, err := c.client.ServiceInspectWithRaw(c.ctx, serviceID, types.ServiceInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect service %s: %w", serviceID, err)
	}
	return &service, nil
}

// CreateService creates a new swarm service
func (c *Client) CreateService(spec swarm.ServiceSpec) (swarm.ServiceCreateResponse, error) {
	resp, err := c.client.ServiceCreate(c.ctx, spec, types.ServiceCreateOptions{})
	if err != nil {
		return swarm.ServiceCreateResponse{}, fmt.Errorf("failed to create service: %w", err)
	}

	logrus.Infof("✅ Service created: %s (%s)", spec.Name, resp.ID)
	return resp, nil
}

// UpdateService updates an existing swarm service
func (c *Client) UpdateService(serviceID string, version swarm.Version, spec swarm.ServiceSpec) error {
	_, err := c.client.ServiceUpdate(c.ctx, serviceID, version, spec, types.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update service %s: %w", serviceID, err)
	}

	logrus.Infof("✅ Service updated: %s", serviceID)
	return nil
}

// RemoveService removes a swarm service
func (c *Client) RemoveService(serviceID string) error {
	err := c.client.ServiceRemove(c.ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to remove service %s: %w", serviceID, err)
	}

	logrus.Infof("✅ Service removed: %s", serviceID)
	return nil
}

// ListTasks returns tasks for a specific service
func (c *Client) ListTasks(serviceID string) ([]swarm.Task, error) {
	filters := types.TaskListOptions{
		Filters: filters.NewArgs(filters.Arg("service", serviceID)),
	}

	tasks, err := c.client.TaskList(c.ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks for service %s: %w", serviceID, err)
	}

	return tasks, nil
}

// ListNodes returns all swarm nodes
func (c *Client) ListNodes() ([]swarm.Node, error) {
	nodes, err := c.client.NodeList(c.ctx, types.NodeListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return nodes, nil
}

// GetNode returns detailed information about a specific node
func (c *Client) GetNode(nodeID string) (*swarm.Node, error) {
	node, _, err := c.client.NodeInspectWithRaw(c.ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect node %s: %w", nodeID, err)
	}
	return &node, nil
}

// UpdateNode updates a swarm node
func (c *Client) UpdateNode(nodeID string, version swarm.Version, spec swarm.NodeSpec) error {
	err := c.client.NodeUpdate(c.ctx, nodeID, version, spec)
	if err != nil {
		return fmt.Errorf("failed to update node %s: %w", nodeID, err)
	}

	logrus.Infof("✅ Node updated: %s", nodeID)
	return nil
}

// RemoveNode removes a swarm node
func (c *Client) RemoveNode(nodeID string, force bool) error {
	err := c.client.NodeRemove(c.ctx, nodeID, types.NodeRemoveOptions{Force: force})
	if err != nil {
		return fmt.Errorf("failed to remove node %s: %w", nodeID, err)
	}

	logrus.Infof("✅ Node removed: %s", nodeID)
	return nil
}
