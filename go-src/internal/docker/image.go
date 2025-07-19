package docker

import (
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/sirupsen/logrus"
)

// ImageInfo represents image information
type ImageInfo struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repo_tags"`
	RepoDigests []string          `json:"repo_digests"`
	Created     int64             `json:"created"`
	Size        int64             `json:"size"`
	Labels      map[string]string `json:"labels"`
}

// ListImages returns all images
func (c *Client) ListImages() ([]ImageInfo, error) {
	images, err := c.client.ImageList(c.ctx, types.ImageListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	var imageInfos []ImageInfo
	for _, img := range images {
		info := ImageInfo{
			ID:          img.ID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Created:     img.Created,
			Size:        img.Size,
			Labels:      img.Labels,
		}
		imageInfos = append(imageInfos, info)
	}

	return imageInfos, nil
}

// GetImage returns detailed information about a specific image
func (c *Client) GetImage(imageID string) (*types.ImageInspect, error) {
	image, _, err := c.client.ImageInspectWithRaw(c.ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image %s: %w", imageID, err)
	}
	return &image, nil
}

// PullImage pulls an image from a registry
func (c *Client) PullImage(imageName string) error {
	reader, err := c.client.ImagePull(c.ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read the response to completion
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to read pull response: %w", err)
	}

	logrus.Infof("✅ Image pulled successfully: %s", imageName)
	return nil
}

// BuildImage builds an image from a Dockerfile
func (c *Client) BuildImage(buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	resp, err := c.client.ImageBuild(c.ctx, buildContext, options)
	if err != nil {
		return types.ImageBuildResponse{}, fmt.Errorf("failed to build image: %w", err)
	}

	logrus.Infof("✅ Image build started: %s", options.Tags[0])
	return resp, nil
}

// RemoveImage removes an image
func (c *Client) RemoveImage(imageID string, force bool) error {
	_, err := c.client.ImageRemove(c.ctx, imageID, types.ImageRemoveOptions{
		Force: force,
	})
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", imageID, err)
	}

	logrus.Infof("✅ Image removed: %s", imageID)
	return nil
}

// TagImage tags an image
func (c *Client) TagImage(imageID, tag string) error {
	err := c.client.ImageTag(c.ctx, imageID, tag)
	if err != nil {
		return fmt.Errorf("failed to tag image %s with %s: %w", imageID, tag, err)
	}

	logrus.Infof("✅ Image tagged: %s -> %s", imageID, tag)
	return nil
}

// PruneImages removes unused images
func (c *Client) PruneImages(filters filters.Args) (*types.ImagesPruneReport, error) {
	report, err := c.client.ImagesPrune(c.ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to prune images: %w", err)
	}

	logrus.Infof("✅ Pruned %d images, freed %d bytes", len(report.ImagesDeleted), report.SpaceReclaimed)
	return &report, nil
}

// SaveImage saves an image to a tar archive
func (c *Client) SaveImage(imageIDs []string) (io.ReadCloser, error) {
	reader, err := c.client.ImageSave(c.ctx, imageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to save images: %w", err)
	}

	return reader, nil
}

// LoadImage loads an image from a tar archive
func (c *Client) LoadImage(archive io.Reader) error {
	resp, err := c.client.ImageLoad(c.ctx, archive, false)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	logrus.Infof("✅ Image loaded: %s", resp.Body)
	return nil
}
