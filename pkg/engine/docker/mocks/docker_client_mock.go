package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

// MockDockerClient is a mock implementation of the Docker client interface
type MockDockerClient struct {
	mock.Mock
}

// ImageBuild mocks the Docker ImageBuild API
func (m *MockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options interface{}) (interface{}, error) {
	args := m.Called(ctx, buildContext, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

// ImageRemove mocks the Docker ImageRemove API
func (m *MockDockerClient) ImageRemove(ctx context.Context, imageID string, options interface{}) ([]interface{}, error) {
	args := m.Called(ctx, imageID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

// ImagePull mocks the Docker ImagePull API
func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, options interface{}) (io.ReadCloser, error) {
	args := m.Called(ctx, refStr, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// ImageInspectWithRaw mocks the Docker ImageInspectWithRaw API
func (m *MockDockerClient) ImageInspectWithRaw(ctx context.Context, imageID string) (interface{}, []byte, error) {
	args := m.Called(ctx, imageID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0), args.Get(1).([]byte), args.Error(2)
}

// ContainerCreate mocks the Docker ContainerCreate API
func (m *MockDockerClient) ContainerCreate(ctx context.Context, config interface{}, hostConfig interface{}, networkingConfig interface{}, platform interface{}, containerName string) (interface{}, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

// ContainerStart mocks the Docker ContainerStart API
func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options interface{}) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

// ContainerWait mocks the Docker ContainerWait API
func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition interface{}) (<-chan interface{}, <-chan error) {
	args := m.Called(ctx, containerID, condition)
	
	statusCh := make(chan interface{}, 1)
	errCh := make(chan error, 1)
	
	if args.Get(0) != nil {
		statusCh <- args.Get(0)
	}
	if args.Get(1) != nil {
		errCh <- args.Error(1)
	}
	
	close(statusCh)
	close(errCh)
	
	return statusCh, errCh
}

// ContainerLogs mocks the Docker ContainerLogs API
func (m *MockDockerClient) ContainerLogs(ctx context.Context, containerID string, options interface{}) (io.ReadCloser, error) {
	args := m.Called(ctx, containerID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// ContainerRemove mocks the Docker ContainerRemove API
func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options interface{}) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

// CopyFromContainer mocks the Docker CopyFromContainer API
func (m *MockDockerClient) CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, interface{}, error) {
	args := m.Called(ctx, containerID, srcPath)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.Get(1), args.Error(2)
}
