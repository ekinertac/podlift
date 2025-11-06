package setup

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ekinertac/podlift/internal/ssh"
)

func TestConfigureFirewall_UFWNotFound(t *testing.T) {
	mockClient := ssh.NewMockClient()
	mockClient.ExecuteFunc = func(cmd string) (string, error) {
		if strings.Contains(cmd, "which ufw") {
			return "", fmt.Errorf("ufw not found")
		}
		return "", nil
	}

	// Should not error, just skip
	err := ConfigureFirewall(mockClient)
	if err != nil {
		t.Errorf("ConfigureFirewall() should not error when UFW not found, got: %v", err)
	}
}

func TestConfigureFirewall_Success(t *testing.T) {
	mockClient := ssh.NewMockClient()
	portsChecked := []int{}

	mockClient.ExecuteFunc = func(cmd string) (string, error) {
		if strings.Contains(cmd, "which ufw") {
			return "/usr/sbin/ufw", nil
		}
		// Check status - return that ports are NOT open yet
		if strings.Contains(cmd, "ufw status | grep") {
			return "", fmt.Errorf("not found") // Port not in firewall rules
		}
		if strings.Contains(cmd, "ufw allow") {
			// Track which ports were configured
			if strings.Contains(cmd, "22/tcp") {
				portsChecked = append(portsChecked, 22)
			}
			if strings.Contains(cmd, "80/tcp") {
				portsChecked = append(portsChecked, 80)
			}
			if strings.Contains(cmd, "443/tcp") {
				portsChecked = append(portsChecked, 443)
			}
			return "", nil
		}
		if strings.Contains(cmd, "ufw --force enable") {
			return "", nil
		}
		return "", nil
	}

	err := ConfigureFirewall(mockClient)
	if err != nil {
		t.Errorf("ConfigureFirewall() error = %v", err)
	}

	if len(portsChecked) != 3 {
		t.Errorf("Expected 3 ports to be configured, got %d", len(portsChecked))
	}
}

func TestCheckFirewall(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn string
		mockError  error
		wantActive bool
		wantErr    bool
	}{
		{
			name:       "firewall active",
			mockReturn: "Status: active\n22/tcp ALLOW",
			mockError:  nil,
			wantActive: true,
			wantErr:    false,
		},
		{
			name:       "firewall inactive",
			mockReturn: "Status: inactive",
			mockError:  nil,
			wantActive: false,
			wantErr:    false,
		},
		{
			name:       "ufw not installed",
			mockReturn: "",
			mockError:  fmt.Errorf("command not found"),
			wantActive: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := ssh.NewMockClient()
			mockClient.ExecuteFunc = func(cmd string) (string, error) {
				return tt.mockReturn, tt.mockError
			}

			active, err := CheckFirewall(mockClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckFirewall() error = %v, wantErr %v", err, tt.wantErr)
			}
			if active != tt.wantActive {
				t.Errorf("CheckFirewall() active = %v, want %v", active, tt.wantActive)
			}
		})
	}
}

