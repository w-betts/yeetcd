package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
)

// MockDockerClient is a mock implementation of the DockerClient interface
type MockDockerClient struct {
	mock.Mock
}

// ImageBuild mocks the Docker ImageBuild API
func (m *MockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	args := m.Called(ctx, buildContext, options)
	if args.Get(0) == nil {
		return types.ImageBuildResponse{}, args.Error(1)
	}
	return args.Get(0).(types.ImageBuildResponse), args.Error(1)
}

// ImageRemove mocks the Docker ImageRemove API
func (m *MockDockerClient) ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	args := m.Called(ctx, imageID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]image.DeleteResponse), args.Error(1)
}

// ImagePull mocks the Docker ImagePull API
func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, refStr, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// ImageInspectWithRaw mocks the Docker ImageInspectWithRaw API
func (m *MockDockerClient) ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, []byte, error) {
	args := m.Called(ctx, imageID)
	if args.Get(0) == nil {
		return types.ImageInspect{}, nil, args.Error(2)
	}
	return args.Get(0).(types.ImageInspect), args.Get(1).([]byte), args.Error(2)
}

// ImageList mocks the Docker ImageList API
func (m *MockDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]image.Summary), args.Error(1)
}

// ContainerCreate mocks the Docker ContainerCreate API
func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	if args.Get(0) == nil {
		return container.CreateResponse{}, args.Error(1)
	}
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

// ContainerStart mocks the Docker ContainerStart API
func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

// ContainerWait mocks the Docker ContainerWait API
func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	args := m.Called(ctx, containerID, condition)

	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)

	if args.Get(0) != nil {
		statusCh <- args.Get(0).(container.WaitResponse)
	}
	if args.Get(1) != nil {
		errCh <- args.Error(1)
	}

	close(statusCh)
	close(errCh)

	return statusCh, errCh
}

// ContainerLogs mocks the Docker ContainerLogs API
func (m *MockDockerClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, containerID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// ContainerRemove mocks the Docker ContainerRemove API
func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

// ContainerInspect mocks the Docker ContainerInspect API
func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return types.ContainerJSON{}, args.Error(1)
	}
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

// CopyFromContainer mocks the Docker CopyFromContainer API
func (m *MockDockerClient) CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error) {
	args := m.Called(ctx, containerID, srcPath)
	if args.Get(0) == nil {
		return nil, container.PathStat{}, args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.Get(1).(container.PathStat), args.Error(2)
}
