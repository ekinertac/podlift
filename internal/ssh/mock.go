package ssh

import (
	"io"
	"strings"
)

// MockClient is a mock SSH client for testing
type MockClient struct {
	ExecuteFunc             func(string) (string, error)
	ExecuteWithOutputFunc   func(string, io.Writer, io.Writer) error
	TestConnectionFunc      func() error
	CheckDockerFunc         func() (string, error)
	CheckPortFunc           func(int) (bool, error)
	GetDiskSpaceFunc        func() (float64, error)
	CheckExistingServiceFunc func(string) (bool, *ServiceInfo, error)
	ListPodliftServicesFunc  func() ([]string, error)
	CompareGitRepoFunc      func(string, string) (bool, error)
	CopyFileFunc            func(string, string) error
	CopyFileWithProgressFunc func(string, string, func(int64, int64)) error
	CloseFunc               func() error
	ConnectFunc             func() error
	Connected               bool
}

func (m *MockClient) Execute(cmd string) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(cmd)
	}
	return "", nil
}

func (m *MockClient) ExecuteWithOutput(cmd string, stdout, stderr io.Writer) error {
	if m.ExecuteWithOutputFunc != nil {
		return m.ExecuteWithOutputFunc(cmd, stdout, stderr)
	}
	return nil
}

func (m *MockClient) TestConnection() error {
	if m.TestConnectionFunc != nil {
		return m.TestConnectionFunc()
	}
	return nil
}

func (m *MockClient) CheckDocker() (string, error) {
	if m.CheckDockerFunc != nil {
		return m.CheckDockerFunc()
	}
	return "Docker version 24.0.5", nil
}

func (m *MockClient) CheckPort(port int) (bool, error) {
	if m.CheckPortFunc != nil {
		return m.CheckPortFunc(port)
	}
	return true, nil
}

func (m *MockClient) GetDiskSpace() (float64, error) {
	if m.GetDiskSpaceFunc != nil {
		return m.GetDiskSpaceFunc()
	}
	return 50.0, nil
}

func (m *MockClient) CheckExistingService(serviceName string) (bool, *ServiceInfo, error) {
	if m.CheckExistingServiceFunc != nil {
		return m.CheckExistingServiceFunc(serviceName)
	}
	return false, nil, nil
}

func (m *MockClient) ListPodliftServices() ([]string, error) {
	if m.ListPodliftServicesFunc != nil {
		return m.ListPodliftServicesFunc()
	}
	return []string{}, nil
}

func (m *MockClient) CompareGitRepo(serviceName, currentRepo string) (bool, error) {
	if m.CompareGitRepoFunc != nil {
		return m.CompareGitRepoFunc(serviceName, currentRepo)
	}
	return true, nil
}

func (m *MockClient) CopyFile(localPath, remotePath string) error {
	if m.CopyFileFunc != nil {
		return m.CopyFileFunc(localPath, remotePath)
	}
	return nil
}

func (m *MockClient) CopyFileWithProgress(localPath, remotePath string, progressFn func(int64, int64)) error {
	if m.CopyFileWithProgressFunc != nil {
		return m.CopyFileWithProgressFunc(localPath, remotePath, progressFn)
	}
	// Simulate progress
	if progressFn != nil {
		progressFn(100, 100)
	}
	return nil
}

func (m *MockClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockClient) Connect() error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc()
	}
	m.Connected = true
	return nil
}

func (m *MockClient) GetStateFile(serviceName string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// NewMockClient creates a mock client with default success responses
func NewMockClient() *MockClient {
	return &MockClient{
		ExecuteFunc: func(cmd string) (string, error) {
			// Return sensible defaults for common commands
			if strings.Contains(cmd, "docker --version") {
				return "Docker version 24.0.5, build abc123", nil
			}
			if strings.Contains(cmd, "docker ps") {
				return "", nil
			}
			if strings.Contains(cmd, "df -BG") {
				return "50G", nil
			}
			return "", nil
		},
		TestConnectionFunc: func() error {
			return nil
		},
		CheckDockerFunc: func() (string, error) {
			return "Docker version 24.0.5", nil
		},
	}
}

