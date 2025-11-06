package ssl

import (
	"fmt"
	"strings"

	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// CertbotManager manages Let's Encrypt certificates
type CertbotManager struct {
	client ssh.SSHClient
}

// NewCertbotManager creates a new certbot manager
func NewCertbotManager(client ssh.SSHClient) *CertbotManager {
	return &CertbotManager{client: client}
}

// IsInstalled checks if certbot is installed
func (m *CertbotManager) IsInstalled() (bool, error) {
	_, err := m.client.Execute("which certbot")
	return err == nil, nil
}

// Install installs certbot
func (m *CertbotManager) Install() error {
	installed, err := m.IsInstalled()
	if err != nil {
		return err
	}

	if installed {
		fmt.Println(ui.Success("certbot already installed"))
		return nil
	}

	fmt.Println(ui.Info("Installing certbot..."))

	// Install certbot via snap (recommended method)
	installCmds := []string{
		"sudo apt-get update",
		"sudo apt-get install -y snapd",
		"sudo snap install core",
		"sudo snap refresh core",
		"sudo snap install --classic certbot",
		"sudo ln -sf /snap/bin/certbot /usr/bin/certbot",
	}

	for _, cmd := range installCmds {
		if _, err := m.client.Execute(cmd); err != nil {
			return fmt.Errorf("certbot installation failed at step '%s': %w", cmd, err)
		}
	}

	fmt.Println(ui.Success("certbot installed"))
	return nil
}

// ObtainCertificate obtains a certificate for a domain
func (m *CertbotManager) ObtainCertificate(domain, email string) error {
	fmt.Println(ui.Info(fmt.Sprintf("Obtaining SSL certificate for %s...", domain)))

	// Use webroot method (nginx must be running)
	cmd := fmt.Sprintf(
		"sudo certbot certonly --webroot -w /var/www/html -d %s --email %s --agree-tos --non-interactive",
		domain, email,
	)

	output, err := m.client.Execute(cmd)
	if err != nil {
		return fmt.Errorf("certificate generation failed: %w\nOutput: %s", err, output)
	}

	fmt.Println(ui.Success(fmt.Sprintf("Certificate obtained for %s", domain)))
	return nil
}

// GetCertificatePath returns the certificate file path
func (m *CertbotManager) GetCertificatePath(domain string) string {
	return fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
}

// GetKeyPath returns the private key file path
func (m *CertbotManager) GetKeyPath(domain string) string {
	return fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", domain)
}

// CheckCertificate checks if certificate exists for domain
func (m *CertbotManager) CheckCertificate(domain string) (bool, error) {
	certPath := m.GetCertificatePath(domain)
	_, err := m.client.Execute(fmt.Sprintf("sudo test -f %s", certPath))
	return err == nil, nil
}

// RenewCertificates renews all certificates
func (m *CertbotManager) RenewCertificates() error {
	fmt.Println(ui.Info("Renewing certificates..."))

	output, err := m.client.Execute("sudo certbot renew --quiet")
	if err != nil {
		return fmt.Errorf("certificate renewal failed: %w\nOutput: %s", err, output)
	}

	fmt.Println(ui.Success("Certificates renewed"))
	return nil
}

// GetCertificateInfo gets information about a certificate
func (m *CertbotManager) GetCertificateInfo(domain string) (map[string]string, error) {
	cmd := fmt.Sprintf("sudo certbot certificates -d %s 2>&1", domain)
	output, err := m.client.Execute(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate info: %w", err)
	}

	info := make(map[string]string)
	
	// Parse output (simple parsing)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Expiry Date:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info["expiry"] = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "Certificate Path:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info["cert_path"] = strings.TrimSpace(parts[1])
			}
		}
	}

	return info, nil
}

// SetupAutoRenewal configures automatic certificate renewal
func (m *CertbotManager) SetupAutoRenewal() error {
	fmt.Println(ui.Info("Setting up automatic renewal..."))

	// certbot snap includes automatic renewal by default
	// Just verify the timer is enabled
	_, err := m.client.Execute("sudo systemctl status snap.certbot.renew.timer")
	if err != nil {
		fmt.Println(ui.Warning("Auto-renewal timer not found (may already be configured)"))
	} else {
		fmt.Println(ui.Success("Auto-renewal configured (runs twice daily)"))
	}

	return nil
}

