package docker

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/sirupsen/logrus"
)

// NetworkInfo represents network information
type NetworkInfo struct {
	ID         string                            `json:"id"`
	Name       string                            `json:"name"`
	Driver     string                            `json:"driver"`
	Scope      string                            `json:"scope"`
	IPAM       network.IPAM                      `json:"ipam"`
	Internal   bool                              `json:"internal"`
	Attachable bool                              `json:"attachable"`
	Labels     map[string]string                 `json:"labels"`
	Containers map[string]types.EndpointResource `json:"containers"`
}

// ListNetworks returns all networks
func (c *Client) ListNetworks() ([]NetworkInfo, error) {
	networks, err := c.client.NetworkList(c.ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	var networkInfos []NetworkInfo
	for _, net := range networks {
		info := NetworkInfo{
			ID:         net.ID,
			Name:       net.Name,
			Driver:     net.Driver,
			Scope:      net.Scope,
			IPAM:       net.IPAM,
			Internal:   net.Internal,
			Attachable: net.Attachable,
			Labels:     net.Labels,
			Containers: net.Containers,
		}
		networkInfos = append(networkInfos, info)
	}

	return networkInfos, nil
}

// GetNetwork returns detailed information about a specific network
func (c *Client) GetNetwork(networkID string) (*types.NetworkResource, error) {
	network, err := c.client.NetworkInspect(c.ctx, networkID, types.NetworkInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect network %s: %w", networkID, err)
	}
	return &network, nil
}

// CreateNetwork creates a new network
func (c *Client) CreateNetwork(name, driver string, options types.NetworkCreate) (types.NetworkCreateResponse, error) {
	if options.Driver == "" {
		options.Driver = driver
	}

	resp, err := c.client.NetworkCreate(c.ctx, name, options)
	if err != nil {
		return types.NetworkCreateResponse{}, fmt.Errorf("failed to create network %s: %w", name, err)
	}

	logrus.Infof("✅ Network created: %s (%s)", name, resp.ID)
	return resp, nil
}

// RemoveNetwork removes a network
func (c *Client) RemoveNetwork(networkID string) error {
	err := c.client.NetworkRemove(c.ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to remove network %s: %w", networkID, err)
	}

	logrus.Infof("✅ Network removed: %s", networkID)
	return nil
}

// ConnectContainerToNetwork connects a container to a network
func (c *Client) ConnectContainerToNetwork(containerID, networkID string, endpointConfig *network.EndpointSettings) error {
	err := c.client.NetworkConnect(c.ctx, networkID, containerID, endpointConfig)
	if err != nil {
		return fmt.Errorf("failed to connect container %s to network %s: %w", containerID, networkID, err)
	}

	logrus.Infof("✅ Container %s connected to network %s", containerID, networkID)
	return nil
}

// DisconnectContainerFromNetwork disconnects a container from a network
func (c *Client) DisconnectContainerFromNetwork(containerID, networkID string, force bool) error {
	err := c.client.NetworkDisconnect(c.ctx, networkID, containerID, force)
	if err != nil {
		return fmt.Errorf("failed to disconnect container %s from network %s: %w", containerID, networkID, err)
	}

	logrus.Infof("✅ Container %s disconnected from network %s", containerID, networkID)
	return nil
}

// CreateOverlayNetwork creates an overlay network for Swarm mode
func (c *Client) CreateOverlayNetwork(name string, options types.NetworkCreate) (types.NetworkCreateResponse, error) {
	options.Driver = "overlay"
	options.Attachable = true

	return c.CreateNetwork(name, "overlay", options)
}

// CreateBridgeNetwork creates a bridge network
func (c *Client) CreateBridgeNetwork(name string, subnet string) (types.NetworkCreateResponse, error) {
	options := types.NetworkCreate{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet: subnet,
				},
			},
		},
		Attachable: true,
	}

	return c.CreateNetwork(name, "bridge", options)
}
