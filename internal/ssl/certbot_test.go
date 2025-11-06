package ssl

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ekinertac/podlift/internal/ssh"
)

func TestGetCertificatePath(t *testing.T) {
	mgr := &CertbotManager{}
	
	path := mgr.GetCertificatePath("example.com")
	expected := "/etc/letsencrypt/live/example.com/fullchain.pem"
	
	if path != expected {
		t.Errorf("GetCertificatePath() = %v, want %v", path, expected)
	}
}

func TestGetKeyPath(t *testing.T) {
	mgr := &CertbotManager{}
	
	path := mgr.GetKeyPath("example.com")
	expected := "/etc/letsencrypt/live/example.com/privkey.pem"
	
	if path != expected {
		t.Errorf("GetKeyPath() = %v, want %v", path, expected)
	}
}

func TestIsInstalled_NotInstalled(t *testing.T) {
	mockClient := ssh.NewMockClient()
	mockClient.ExecuteFunc = func(cmd string) (string, error) {
		if strings.Contains(cmd, "which certbot") {
			return "", fmt.Errorf("not found")
		}
		return "", nil
	}

	mgr := NewCertbotManager(mockClient)
	installed, err := mgr.IsInstalled()
	
	if err != nil {
		t.Errorf("IsInstalled() error = %v", err)
	}
	if installed {
		t.Error("IsInstalled() should return false when certbot not found")
	}
}

func TestIsInstalled_Installed(t *testing.T) {
	mockClient := ssh.NewMockClient()
	mockClient.ExecuteFunc = func(cmd string) (string, error) {
		if strings.Contains(cmd, "which certbot") {
			return "/usr/bin/certbot", nil
		}
		return "", nil
	}

	mgr := NewCertbotManager(mockClient)
	installed, err := mgr.IsInstalled()
	
	if err != nil {
		t.Errorf("IsInstalled() error = %v", err)
	}
	if !installed {
		t.Error("IsInstalled() should return true when certbot found")
	}
}

func TestCheckCertificate(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		fileExists bool
		want       bool
	}{
		{
			name:       "certificate exists",
			domain:     "example.com",
			fileExists: true,
			want:       true,
		},
		{
			name:       "certificate missing",
			domain:     "missing.com",
			fileExists: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := ssh.NewMockClient()
			mockClient.ExecuteFunc = func(cmd string) (string, error) {
				if strings.Contains(cmd, "test -f") {
					if tt.fileExists {
						return "", nil
					}
					return "", fmt.Errorf("file not found")
				}
				return "", nil
			}

			mgr := NewCertbotManager(mockClient)
			exists, err := mgr.CheckCertificate(tt.domain)
			
			if err != nil {
				t.Errorf("CheckCertificate() error = %v", err)
			}
			if exists != tt.want {
				t.Errorf("CheckCertificate() = %v, want %v", exists, tt.want)
			}
		})
	}
}

