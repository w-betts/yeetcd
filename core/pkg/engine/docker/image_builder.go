package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// DockerDaemonImageBuilder builds Docker images using the Docker daemon
type DockerDaemonImageBuilder struct {
	logger *slog.Logger
}

// NewDockerDaemonImageBuilder creates a new image builder
func NewDockerDaemonImageBuilder() *DockerDaemonImageBuilder {
	return &DockerDaemonImageBuilder{
		logger: slog.Default().With("component", "image-builder"),
	}
}

// BuildImage builds a Docker image using the Docker daemon API
func (b *DockerDaemonImageBuilder) BuildImage(ctx context.Context, dockerClient DockerClient, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
	b.logger.Info("building image", "image", def.Image, "tag", def.Tag)

	// Create Dockerfile
	dockerfile, cleanup, err := CreateDockerfile(ctx, def, def.ArtifactDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create Dockerfile: %w", err)
	}
	defer cleanup()

	b.logger.Debug("created Dockerfile", "path", dockerfile)

	// Build the image
	imageTag := fmt.Sprintf("%s:%s", def.Image, def.Tag)

	// Create build context (tar archive of the artifact directory)
	buildContext, err := createBuildContext(def.ArtifactDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	// Build the image
	// Dockerfile path should be relative to the build context
	resp, err := dockerClient.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageTag},
		Remove:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	// Wait for build to complete and get image ID
	imageID, err := waitForBuildCompletion(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error during image build: %w", err)
	}

	b.logger.Info("image built successfully", "imageID", imageID)
	return &engine.BuildImageResult{ImageID: imageID}, nil
}

// createBuildContext creates a tar archive of the given directory
func createBuildContext(dir string) (io.ReadCloser, error) {
	// Create a pipe
	pr, pw := io.Pipe()

	// Create tar writer
	tw := tar.NewWriter(pw)

	// Write directory contents to tar in a goroutine
	go func() {
		defer pw.Close()
		defer tw.Close()

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = relPath

			// Write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// Write file contents if it's a regular file
			if info.Mode().IsRegular() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				if _, err := io.Copy(tw, file); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

// waitForBuildCompletion waits for the Docker build to complete and returns the image ID
func waitForBuildCompletion(reader io.Reader) (string, error) {
	decoder := json.NewDecoder(reader)
	var imageID string

	for {
		var msg jsonmessage.JSONMessage
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error reading build output: %w", err)
		}

		// Check for error messages
		if msg.Error != nil {
			return "", fmt.Errorf("build error: %s", msg.Error.Message)
		}

		// Extract image ID from the aux message
		if msg.Aux != nil {
			// The aux message contains build details
			// Parse it to get the image ID
			var auxData map[string]interface{}
			if err := json.Unmarshal(*msg.Aux, &auxData); err == nil {
				if id, ok := auxData["ID"].(string); ok {
					imageID = id
				}
			}
		}
	}

	// If we didn't get an image ID from aux, we need to parse the stream
	// The image ID is typically in the format "sha256:..."
	if imageID == "" {
		// Read all output and parse for image ID
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, reader); err != nil {
			return "", fmt.Errorf("error reading build output: %w", err)
		}

		// Parse the output for "Successfully built <image-id>"
		output := buf.String()
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Successfully built") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					imageID = parts[2]
					break
				}
			}
		}
	}

	if imageID == "" {
		return "", fmt.Errorf("failed to get image ID from build output")
	}

	return imageID, nil
}
