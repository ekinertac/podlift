package setup

import (
	"fmt"
	"io"
	"testing"

	"github.com/ekinertac/podlift/internal/ssh"
)

func TestInstallDocker_AlreadyInstalled(t *testing.T) {
	mockClient := ssh.NewMockClient()
	mockClient.CheckDockerFunc = func() (string, error) {
		return "Docker version 24.0.5", nil
	}

	err := InstallDocker(mockClient)
	if err != nil {
		t.Errorf("InstallDocker() error = %v, want nil when already installed", err)
	}
}

func TestInstallDocker_InstallationFails(t *testing.T) {
	mockClient := ssh.NewMockClient()
	mockClient.CheckDockerFunc = func() (string, error) {
		return "", fmt.Errorf("docker not found")
	}
	mockClient.ExecuteWithOutputFunc = func(cmd string, stdout, stderr io.Writer) error {
		return fmt.Errorf("installation failed")
	}

	err := InstallDocker(mockClient)
	if err == nil {
		t.Error("InstallDocker() should fail when installation fails")
	}
}

func TestCheckDockerVersion_Success(t *testing.T) {
	mockClient := ssh.NewMockClient()
	
	version, err := CheckDockerVersion(mockClient)
	if err != nil {
		t.Errorf("CheckDockerVersion() error = %v", err)
	}

	if version == "" {
		t.Error("CheckDockerVersion() should return version string")
	}
}

