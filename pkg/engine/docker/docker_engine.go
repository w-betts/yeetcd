package docker

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// DockerExecutionEngine implements the ExecutionEngine interface using Docker
type DockerExecutionEngine struct {
	dockerClient DockerClient
	imageBuilder *DockerDaemonImageBuilder
	runner       *ContainerRunner
	logger       *slog.Logger
}

// NewDockerExecutionEngine creates a new Docker execution engine
func NewDockerExecutionEngine() (*DockerExecutionEngine, error) {
	cli, err := NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	logger := slog.Default().With("component", "docker-engine")

	return &DockerExecutionEngine{
		dockerClient: cli,
		imageBuilder: NewDockerDaemonImageBuilder(),
		runner:       NewContainerRunner(),
		logger:       logger,
	}, nil
}

// NewDockerExecutionEngineWithClient creates a new Docker execution engine with a specific client (for testing)
func NewDockerExecutionEngineWithClient(cli DockerClient) *DockerExecutionEngine {
	logger := slog.Default().With("component", "docker-engine")

	return &DockerExecutionEngine{
		dockerClient: cli,
		imageBuilder: NewDockerDaemonImageBuilder(),
		runner:       NewContainerRunner(),
		logger:       logger,
	}
}

// BuildImage builds a Docker image from the given definition
func (d *DockerExecutionEngine) BuildImage(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
	d.logger.Info("building image", "image", def.Image, "tag", def.Tag)
	return d.imageBuilder.BuildImage(ctx, d.dockerClient, def)
}

// RemoveImage removes a Docker image
func (d *DockerExecutionEngine) RemoveImage(ctx context.Context, imageID string) error {
	d.logger.Info("removing image", "imageID", imageID)
	_, err := d.dockerClient.ImageRemove(ctx, imageID, image.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", imageID, err)
	}
	return nil
}

// RunJob runs a job in a Docker container
func (d *DockerExecutionEngine) RunJob(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
	d.logger.Info("running job", "image", def.Image)

	// Create temp directory for output
	outputDir, err := os.MkdirTemp("", "yeetcd-output-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	var containerID string
	exitCode := -1

	// Ensure cleanup
	defer func() {
		if containerID != "" {
			if err := d.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{
				Force: true,
			}); err != nil {
				d.logger.Warn("failed to remove container", "containerID", containerID, "error", err)
			}
		}
	}()

	// Pull image if not present
	if err := d.runner.PullImage(ctx, d.dockerClient, def.Image); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Create container
	containerID, err = d.runner.CreateContainer(ctx, d.dockerClient, def)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Run container
	exitCode, err = d.runner.RunContainer(ctx, d.dockerClient, containerID, def.JobStreams)
	if err != nil {
		return &engine.JobResult{ExitCode: exitCode, OutputDirectoriesParent: outputDir}, fmt.Errorf("failed to run container: %w", err)
	}

	// Extract output directories if job succeeded
	if exitCode == 0 && len(def.OutputDirectoryPaths) > 0 {
		for name, path := range def.OutputDirectoryPaths {
			if err := d.extractOutputDirectory(ctx, containerID, outputDir, name, path); err != nil {
				return &engine.JobResult{ExitCode: exitCode, OutputDirectoriesParent: outputDir}, fmt.Errorf("failed to extract output directory: %w", err)
			}
		}
	}

	d.logger.Info("job completed", "exitCode", exitCode)
	return &engine.JobResult{ExitCode: exitCode, OutputDirectoriesParent: outputDir}, nil
}

// extractOutputDirectory extracts files from a container path to the output directory
func (d *DockerExecutionEngine) extractOutputDirectory(ctx context.Context, containerID, outputDir, name, path string) error {
	d.logger.Debug("extracting output directory", "containerID", containerID, "name", name, "path", path)

	// Create a temp directory for extraction
	tempDir, err := os.MkdirTemp("", "yeetcd-extract-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy archive from container
	content, stat, err := d.dockerClient.CopyFromContainer(ctx, containerID, path)
	if err != nil {
		return fmt.Errorf("failed to copy from container: %w", err)
	}
	defer content.Close()

	d.logger.Debug("extracted archive", "path", path, "stat", stat)

	// Extract tar archive to temp directory
	if err := extractTarArchive(content, tempDir); err != nil {
		return fmt.Errorf("failed to extract tar archive: %w", err)
	}

	// The tar archive contains the directory itself as the root.
	// For example, if path is "/output", the tar contains "output/result.txt".
	// After extraction to tempDir, we have tempDir/output/result.txt.
	// We need to move tempDir/output to outputDir/name.
	destDir := filepath.Join(outputDir, name)

	// Read the contents of the temp directory to find the extracted directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	// Find the extracted directory (should be the only entry)
	for _, entry := range entries {
		srcPath := filepath.Join(tempDir, entry.Name())
		if err := os.Rename(srcPath, destDir); err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", srcPath, destDir, err)
		}
		break // Only move the first entry (the extracted directory)
	}

	return nil
}

// extractTarArchive extracts a tar archive to the destination directory
func extractTarArchive(reader io.Reader, dest string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", target, err)
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file %s: %w", target, err)
			}
			file.Close()
		}
	}

	return nil
}

// pullImageIfNeeded pulls an image if it's not present locally
func pullImageIfNeeded(ctx context.Context, cli DockerClient, imageRef string) error {
	// Check if image exists
	_, _, err := cli.ImageInspectWithRaw(ctx, imageRef)
	if err == nil {
		// Image exists
		return nil
	}

	// Image doesn't exist, pull it
	out, err := cli.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageRef, err)
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
	}

	return nil
}

// createContainer creates a Docker container with the given configuration
func createContainer(ctx context.Context, cli DockerClient, def engine.JobDefinition) (string, error) {
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

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

// runContainer starts a container and waits for it to complete
func runContainer(ctx context.Context, cli DockerClient, containerID string, streams *engine.JobStreams) (int, error) {
	// Start container
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return -1, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return -1, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		// Container finished
		if status.StatusCode != 0 {
			// Get logs even if failed
		}
	}

	// Get container logs
	logs, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	// Copy logs to streams
	if streams != nil {
		if stdout := streams.StdoutWriter(); stdout != nil {
			io.Copy(stdout, logs)
		}
	}

	// Get exit code
	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return -1, fmt.Errorf("failed to inspect container: %w", err)
	}

	return inspect.State.ExitCode, nil
}

// extractOutputFiles extracts files from a container path
func extractOutputFiles(ctx context.Context, cli DockerClient, containerID, path string) ([]byte, error) {
	content, _, err := cli.CopyFromContainer(ctx, containerID, path)
	if err != nil {
		return nil, fmt.Errorf("failed to copy from container: %w", err)
	}
	defer content.Close()

	data, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	return data, nil
}

// listImages lists all Docker images
func listImages(ctx context.Context, cli DockerClient) ([]image.Summary, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	return images, nil
}

// findImageByTag finds an image by its tag
func findImageByTag(ctx context.Context, cli DockerClient, tag string) (*image.Summary, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: tag,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("image not found: %s", tag)
	}

	return &images[0], nil
}
