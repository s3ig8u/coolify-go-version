package docker

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListImages(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	images, err := client.ListImages()
	assert.NoError(t, err)
	assert.NotNil(t, images)
}

func TestClient_GetImage(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// First, list images to get an existing image ID
	images, err := client.ListImages()
	if err != nil {
		t.Skip("No images available for testing")
	}

	if len(images) == 0 {
		t.Skip("No images available for testing")
	}

	imageID := images[0].ID
	image, err := client.GetImage(imageID)
	assert.NoError(t, err)
	assert.NotNil(t, image)
	assert.Equal(t, imageID, image.ID)
}

func TestClient_PullImage(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Pull a small test image
	imageName := "alpine:latest"
	err = client.PullImage(imageName)
	assert.NoError(t, err)

	// Verify the image was pulled by listing images
	images, err := client.ListImages()
	assert.NoError(t, err)

	// Check if the pulled image exists
	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	assert.True(t, found, "Pulled image should be in the list")
}

func TestClient_BuildImage(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Create a simple Dockerfile content
	dockerfile := `FROM alpine:latest
RUN echo "hello world" > /test.txt
CMD ["cat", "/test.txt"]`

	// Create build context
	buildContext := strings.NewReader(dockerfile)

	// Build the image
	imageName := "test-build-image"
	options := types.ImageBuildOptions{
		Tags: []string{imageName},
	}

	resp, err := client.BuildImage(buildContext, options)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Body)
	defer resp.Body.Close()

	// Verify the image was built
	images, err := client.ListImages()
	assert.NoError(t, err)

	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	assert.True(t, found, "Built image should be in the list")

	// Clean up
	defer func() {
		client.RemoveImage(imageName, true)
	}()
}

func TestClient_RemoveImage(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// First, pull an image to remove
	imageName := "busybox:latest"
	err = client.PullImage(imageName)
	require.NoError(t, err)

	// Verify the image exists
	images, err := client.ListImages()
	require.NoError(t, err)

	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	require.True(t, found, "Image should exist before removal")

	// Remove the image
	err = client.RemoveImage(imageName, false)
	assert.NoError(t, err)

	// Verify the image was removed
	images, err = client.ListImages()
	assert.NoError(t, err)

	found = false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	assert.False(t, found, "Image should not exist after removal")
}

func TestClient_TagImage(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Pull a test image
	imageName := "alpine:latest"
	err = client.PullImage(imageName)
	require.NoError(t, err)

	// Tag the image
	newTag := "alpine:test-tag"
	err = client.TagImage(imageName, newTag)
	assert.NoError(t, err)

	// Verify the tagged image exists
	images, err := client.ListImages()
	assert.NoError(t, err)

	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == newTag {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	assert.True(t, found, "Tagged image should exist")

	// Clean up
	client.RemoveImage(newTag, false)
}

func TestClient_SaveImage(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Pull a small test image
	imageName := "alpine:latest"
	err = client.PullImage(imageName)
	require.NoError(t, err)

	// Save the image
	imageIDs := []string{imageName}
	saveReader, err := client.SaveImage(imageIDs)
	assert.NoError(t, err)
	assert.NotNil(t, saveReader)
	defer saveReader.Close()
}

func TestClient_LoadImage(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Pull a small test image
	imageName := "alpine:latest"
	err = client.PullImage(imageName)
	require.NoError(t, err)

	// Save the image
	imageIDs := []string{imageName}
	saveReader, err := client.SaveImage(imageIDs)
	require.NoError(t, err)
	defer saveReader.Close()

	// Read the saved image into a buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, saveReader)
	require.NoError(t, err)

	// Load the image back
	loadReader := bytes.NewReader(buf.Bytes())
	err = client.LoadImage(loadReader)
	assert.NoError(t, err)
}

func TestClient_PruneImages(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	// Prune unused images
	report, err := client.PruneImages(filters.Args{})
	assert.NoError(t, err)
	assert.NotNil(t, report)
}
