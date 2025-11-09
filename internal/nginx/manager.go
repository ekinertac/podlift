package nginx

import (
	"fmt"

	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// Manager handles nginx operations
type Manager struct {
	client ssh.SSHClient
}

// NewManager creates a new nginx manager
func NewManager(client ssh.SSHClient) *Manager {
	return &Manager{client: client}
}

// IsInstalled checks if nginx is installed
func (m *Manager) IsInstalled() (bool, error) {
	_, err := m.client.Execute(GenerateCheckCommand())
	return err == nil, nil
}

// Install installs nginx
func (m *Manager) Install() error {
	// Check if already installed
	installed, err := m.IsInstalled()
	if err != nil {
		return err
	}

	if installed {
		fmt.Println(ui.Success("nginx already installed"))
		return nil
	}

	fmt.Println(ui.Info("Installing nginx..."))
	
	_, err = m.client.Execute(GenerateInstallCommand())
	if err != nil {
		return fmt.Errorf("nginx installation failed: %w", err)
	}
	
	// Disable default site to avoid conflicts with our configuration
	m.client.Execute("sudo rm -f /etc/nginx/sites-enabled/default")

	fmt.Println(ui.Success("nginx installed"))
	return nil
}

// WriteConfig writes nginx configuration to server
func (m *Manager) WriteConfig(serviceName, config string) error {
	sitePath := GenerateSitePath(serviceName)
	return m.client.WriteFile(config, sitePath)
}

// EnableSite enables an nginx site
func (m *Manager) EnableSite(serviceName string) error {
	cmd := GenerateEnableCommand(serviceName)
	_, err := m.client.Execute(cmd)
	if err != nil {
		return fmt.Errorf("failed to enable site: %w", err)
	}
	return nil
}

// TestConfig tests nginx configuration
func (m *Manager) TestConfig() error {
	output, err := m.client.Execute(GenerateTestCommand())
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s", output)
	}
	return nil
}

// Reload reloads nginx
func (m *Manager) Reload() error {
	// Test config first
	if err := m.TestConfig(); err != nil {
		return err
	}

	// Reload
	_, err := m.client.Execute(GenerateReloadCommand())
	if err != nil {
		return fmt.Errorf("nginx reload failed: %w", err)
	}

	fmt.Println(ui.Success("nginx reloaded"))
	return nil
}

// UpdateUpstream updates nginx upstream without full reload
func (m *Manager) UpdateUpstream(serviceName string, upstreams []Upstream, domain string, ssl SSLConfig) error {
	// Generate new config
	cfg := Config{
		Domain:      domain,
		ServiceName: serviceName,
		Upstreams:   upstreams,
		SSL:         ssl,
	}

	config, err := GenerateConfig(cfg)
	if err != nil {
		return err
	}

	// Write config
	if err := m.WriteConfig(serviceName, config); err != nil {
		return err
	}

	// Enable site
	if err := m.EnableSite(serviceName); err != nil {
		return err
	}

	// Reload nginx
	if err := m.Reload(); err != nil {
		return err
	}

	return nil
}

// RemoveSite removes nginx site configuration
func (m *Manager) RemoveSite(serviceName string) error {
	sitePath := GenerateSitePath(serviceName)
	symlinkPath := GenerateSymlinkPath(serviceName)

	// Remove symlink
	m.client.Execute(fmt.Sprintf("sudo rm -f %s", symlinkPath))
	
	// Remove config
	m.client.Execute(fmt.Sprintf("sudo rm -f %s", sitePath))

	// Reload
	m.Reload()

	return nil
}

