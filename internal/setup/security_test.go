package setup

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ekinertac/podlift/internal/ssh"
)

func TestApplySecurity(t *testing.T) {
	mockClient := ssh.NewMockClient()
	commandsRun := []string{}

	mockClient.ExecuteFunc = func(cmd string) (string, error) {
		commandsRun = append(commandsRun, cmd)
		
		if strings.Contains(cmd, "which fail2ban-client") {
			return "", fmt.Errorf("not found") // Not installed
		}
		return "", nil
	}

	err := ApplySecurity(mockClient)
	if err != nil {
		t.Errorf("ApplySecurity() error = %v", err)
	}

	// Should have tried to disable password auth
	found := false
	for _, cmd := range commandsRun {
		if strings.Contains(cmd, "PasswordAuthentication") {
			found = true
			break
		}
	}
	if !found {
		t.Error("ApplySecurity() should try to disable password auth")
	}
}

func TestVerifySetup(t *testing.T) {
	mockClient := ssh.NewMockClient()
	mockClient.CheckDockerFunc = func() (string, error) {
		return "Docker version 24.0.5", nil
	}
	mockClient.ExecuteFunc = func(cmd string) (string, error) {
		if strings.Contains(cmd, "ufw status") {
			return "Status: active", nil
		}
		return "", nil
	}
	mockClient.TestConnectionFunc = func() error {
		return nil
	}

	err := VerifySetup(mockClient)
	if err != nil {
		t.Errorf("VerifySetup() error = %v", err)
	}
}

func TestVerifySetup_DockerFailed(t *testing.T) {
	mockClient := ssh.NewMockClient()
	mockClient.CheckDockerFunc = func() (string, error) {
		return "", fmt.Errorf("docker not found")
	}

	err := VerifySetup(mockClient)
	if err == nil {
		t.Error("VerifySetup() should fail when Docker check fails")
	}
}

