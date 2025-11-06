package ssh

import "io"

// SSHClient interface for SSH operations (allows mocking)
type SSHClient interface {
	Connect() error
	Close() error
	Execute(cmd string) (string, error)
	ExecuteWithOutput(cmd string, stdout, stderr io.Writer) error
	TestConnection() error
	CheckDocker() (string, error)
	CheckPort(port int) (bool, error)
	GetDiskSpace() (float64, error)
	CheckExistingService(serviceName string) (bool, *ServiceInfo, error)
	ListPodliftServices() ([]string, error)
	CompareGitRepo(serviceName, currentRepo string) (bool, error)
	CopyFile(localPath, remotePath string) error
	CopyFileWithProgress(localPath, remotePath string, progressFn func(int64, int64)) error
	GetStateFile(serviceName string) (map[string]interface{}, error)
}

// Verify that Client implements SSHClient
var _ SSHClient = (*Client)(nil)

// Verify that MockClient implements SSHClient  
var _ SSHClient = (*MockClient)(nil)

