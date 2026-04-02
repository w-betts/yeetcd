package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// ContainerRunner handles container lifecycle operations
type ContainerRunner struct {
	logger *slog.Logger
}

// NewContainerRunner creates a new container runner
func NewContainerRunner() *ContainerRunner {
	return &ContainerRunner{
		logger: slog.Default().With("component", "container-runner"),
	}
}

// PullImage pulls a Docker image if not present locally
func (r *ContainerRunner) PullImage(ctx context.Context, dockerClient DockerClient, imageTag string) error {
	r.logger.Debug("checking if image exists", "image", imageTag)

	// Check if image exists locally
	_, _, err := dockerClient.ImageInspectWithRaw(ctx, imageTag)
	if err == nil {
		r.logger.Debug("image already exists locally", "image", imageTag)
		return nil
	}

	// Image doesn't exist, pull it
	r.logger.Info("pulling image", "image", imageTag)

	out, err := dockerClient.ImagePull(ctx, imageTag, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageTag, err)
	}
	defer out.Close()

	// Wait for pull to complete
	decoder := json.NewDecoder(out)
	for {
		var msg jsonmessage.JSONMessage
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading pull output: %w", err)
		}

		// Check for error messages
		if msg.Error != nil {
			return fmt.Errorf("pull error: %s", msg.Error.Message)
		}
	}

	r.logger.Info("image pulled successfully", "image", imageTag)
	return nil
}

// CreateContainer creates a new container with the given configuration
func (r *ContainerRunner) CreateContainer(ctx context.Context, dockerClient DockerClient, def engine.JobDefinition) (string, error) {
	r.logger.Debug("creating container", "image", def.Image)

	// Build environment variables
	var env []string
	for k, v := range def.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Build binds for volume mounts
	var binds []string
	for containerPath, mountInput := range def.InputFilePaths {
		binds = append(binds, fmt.Sprintf("%s:%s", mountInput.Directory(), containerPath))
	}

	// Create container
	config := &container.Config{
		Image:      def.Image,
		Cmd:        def.Cmd,
		WorkingDir: def.WorkingDir,
		Env:        env,
	}

	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	resp, err := dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	r.logger.Debug("container created", "containerID", resp.ID)
	return resp.ID, nil
}

// RunContainer runs a container and captures its output
func (r *ContainerRunner) RunContainer(ctx context.Context, dockerClient DockerClient, containerID string, streams *engine.JobStreams) (int, error) {
	r.logger.Debug("starting container", "containerID", containerID)

	// Attach to container's streams before starting (if streams are needed)
	// This is similar to Java's logContainerCmd with withFollowStream(true)
	var attachResp types.HijackedResponse
	var attachSuccessful bool
	if streams != nil {
		var err error
		attachResp, err = dockerClient.ContainerAttach(ctx, containerID, container.AttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
		})
		if err != nil {
			r.logger.Warn("failed to attach to container", "error", err)
		} else if attachResp.Conn != nil {
			attachSuccessful = true
		}
	}
	
	// Only close if attach was successful
	if attachSuccessful {
		defer attachResp.Close()
	}

	// Start container
	if err := dockerClient.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return -1, fmt.Errorf("failed to start container: %w", err)
	}

	// If we have an attach response, demultiplex the stream in a goroutine
	var demultiplexDone chan struct{}
	if attachSuccessful {
		demultiplexDone = make(chan struct{})
		go func() {
			defer close(demultiplexDone)
			stdout := streams.StdoutWriter()
			stderr := streams.StderrWriter()
			if _, err := stdcopy.StdCopy(stdout, stderr, attachResp.Reader); err != nil {
				r.logger.Warn("failed to demultiplex attached stream", "error", err)
			}
		}()
	}

	r.logger.Debug("waiting for container to complete", "containerID", containerID)

	// Wait for container to finish
	statusCh, errCh := dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	var exitCode int

	select {
	case err := <-errCh:
		if err != nil {
			return -1, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	}

	r.logger.Debug("container finished", "containerID", containerID, "exitCode", exitCode)

	// Wait for demultiplexing to complete
	if demultiplexDone != nil {
		<-demultiplexDone
	}

	return exitCode, nil
}

// ExtractArchive extracts files from a container path
func (r *ContainerRunner) ExtractArchive(ctx context.Context, dockerClient DockerClient, containerID string, path string) ([]byte, error) {
	r.logger.Debug("extracting archive", "containerID", containerID, "path", path)

	content, _, err := dockerClient.CopyFromContainer(ctx, containerID, path)
	if err != nil {
		return nil, fmt.Errorf("failed to copy from container: %w", err)
	}
	defer content.Close()

	data, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	r.logger.Debug("archive extracted", "path", path, "size", len(data))
	return data, nil
}

// ExtractArchiveToDir extracts files from a container path to a directory
func (r *ContainerRunner) ExtractArchiveToDir(ctx context.Context, dockerClient DockerClient, containerID string, containerPath string, destDir string) error {
	r.logger.Debug("extracting archive to directory", "containerID", containerID, "path", containerPath, "dest", destDir)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy archive from container
	content, _, err := dockerClient.CopyFromContainer(ctx, containerID, containerPath)
	if err != nil {
		return fmt.Errorf("failed to copy from container: %w", err)
	}
	defer content.Close()

	// Extract tar archive
	if err := extractTarArchive(content, destDir); err != nil {
		return fmt.Errorf("failed to extract tar archive: %w", err)
	}

	r.logger.Debug("archive extracted to directory", "path", containerPath, "dest", destDir)
	return nil
}
