package setup

import (
	"fmt"
	"strings"

	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// ConfigureFirewall configures UFW firewall with necessary ports
func ConfigureFirewall(client *ssh.Client) error {
	// Check if UFW is installed
	_, err := client.Execute("which ufw")
	if err != nil {
		fmt.Println(ui.Warning("UFW not found (skip with --no-firewall)"))
		return nil
	}

	ports := []struct {
		port int
		desc string
	}{
		{22, "SSH"},
		{80, "HTTP"},
		{443, "HTTPS"},
	}

	for _, p := range ports {
		// Check if port is already allowed
		checkCmd := fmt.Sprintf("ufw status | grep '%d/tcp' | grep ALLOW", p.port)
		_, err := client.Execute(checkCmd)
		
		if err == nil {
			fmt.Println(ui.Info(fmt.Sprintf("Port %d (%s) already open", p.port, p.desc)))
			continue
		}

		// Allow port
		fmt.Println(ui.Info(fmt.Sprintf("Opening port %d (%s)...", p.port, p.desc)))
		allowCmd := fmt.Sprintf("ufw allow %d/tcp", p.port)
		if _, err := client.Execute(allowCmd); err != nil {
			return fmt.Errorf("failed to open port %d: %w", p.port, err)
		}
	}

	// Enable firewall (with --force to avoid prompts)
	_, err = client.Execute("ufw --force enable")
	if err != nil {
		return fmt.Errorf("failed to enable firewall: %w", err)
	}

	fmt.Println(ui.Success("Firewall configured"))
	return nil
}

// CheckFirewall checks if firewall is configured
func CheckFirewall(client *ssh.Client) (bool, error) {
	output, err := client.Execute("ufw status")
	if err != nil {
		return false, nil // UFW not installed
	}

	// Check if firewall is active
	if !strings.Contains(output, "Status: active") {
		return false, nil
	}

	return true, nil
}

