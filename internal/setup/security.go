package setup

import (
	"fmt"
	"strings"

	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// ApplySecurity applies basic security hardening
func ApplySecurity(client ssh.SSHClient) error {
	// 1. Disable password authentication for SSH
	fmt.Println(ui.Info("Disabling SSH password authentication..."))
	
	disablePasswordCmd := `sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/g' /etc/ssh/sshd_config && \
sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/g' /etc/ssh/sshd_config && \
systemctl reload sshd || service sshd reload`

	if _, err := client.Execute(disablePasswordCmd); err != nil {
		fmt.Println(ui.Warning("Failed to disable password auth (may already be disabled)"))
	} else {
		fmt.Println(ui.Info("  SSH password auth disabled"))
	}

	// 2. Install fail2ban
	fmt.Println(ui.Info("Installing fail2ban..."))
	
	// Check if already installed
	if _, err := client.Execute("which fail2ban-client"); err == nil {
		fmt.Println(ui.Info("  fail2ban already installed"))
	} else {
		installCmd := "DEBIAN_FRONTEND=noninteractive apt-get install -y fail2ban 2>/dev/null || yum install -y fail2ban 2>/dev/null"
		if _, err := client.Execute(installCmd); err != nil {
			fmt.Println(ui.Warning("Failed to install fail2ban (optional)"))
		} else {
			client.Execute("systemctl enable fail2ban && systemctl start fail2ban")
			fmt.Println(ui.Info("  fail2ban installed and enabled"))
		}
	}

	fmt.Println(ui.Success("Security configured"))
	return nil
}

// VerifySetup verifies that setup was successful
func VerifySetup(client ssh.SSHClient) error {
	checks := []struct {
		name string
		fn   func() error
	}{
		{"Docker running", func() error {
			_, err := client.CheckDocker()
			return err
		}},
		{"Firewall enabled", func() error {
			output, _ := client.Execute("ufw status 2>/dev/null")
			if strings.Contains(output, "Status: active") {
				return nil
			}
			return fmt.Errorf("firewall not active")
		}},
		{"SSH access working", func() error {
			return client.TestConnection()
		}},
	}

	for _, check := range checks {
		if err := check.fn(); err != nil {
			return fmt.Errorf("%s: %w", check.name, err)
		}
		fmt.Println(ui.Success(check.name))
	}

	return nil
}

