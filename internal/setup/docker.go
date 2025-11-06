package setup

import (
	"fmt"
	"strings"

	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// InstallDocker installs Docker on the server if not already installed
func InstallDocker(client ssh.SSHClient) error {
	// Check if Docker is already installed
	version, err := client.CheckDocker()
	if err == nil {
		fmt.Println(ui.Success(fmt.Sprintf("Docker %s already installed", strings.TrimSpace(version))))
		return nil
	}

	// Install Docker using official script
	fmt.Println(ui.Info("Docker not installed, installing..."))

	// Download and run get.docker.com script
	installScript := `curl -fsSL https://get.docker.com | sh && \
sudo systemctl enable docker && \
sudo systemctl start docker && \
sudo usermod -aG docker $USER`

	if err := client.ExecuteWithOutput(installScript, &stdoutWriter{}, &stderrWriter{}); err != nil {
		return fmt.Errorf("Docker installation failed: %w", err)
	}
	
	fmt.Println(ui.Info("  Adding user to docker group (requires logout/login to take effect)"))
	fmt.Println(ui.Info("  Using sudo for docker commands temporarily..."))

	// Verify installation
	version, err = client.CheckDocker()
	if err != nil {
		return fmt.Errorf("Docker installed but not working: %w", err)
	}

	fmt.Println(ui.Success(fmt.Sprintf("Docker %s installed", strings.TrimSpace(version))))
	return nil
}

// CheckDockerVersion checks if Docker version meets minimum requirements
func CheckDockerVersion(client ssh.SSHClient) (string, error) {
	version, err := client.CheckDocker()
	if err != nil {
		return "", err
	}

	// TODO: Parse version and check minimum (20.10+)
	return strings.TrimSpace(version), nil
}

// Simple stdout/stderr writers for command output
type stdoutWriter struct{}

func (w *stdoutWriter) Write(p []byte) (n int, err error) {
	output := strings.TrimSpace(string(p))
	if output != "" {
		fmt.Println(ui.Info("  " + output))
	}
	return len(p), nil
}

type stderrWriter struct{}

func (w *stderrWriter) Write(p []byte) (n int, err error) {
	output := strings.TrimSpace(string(p))
	if output != "" {
		fmt.Println(ui.Warning("  " + output))
	}
	return len(p), nil
}

