package registry

import (
	"testing"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/ssh"
)

func TestGetImagePath(t *testing.T) {
	tests := []struct {
		name      string
		registry  *config.RegistryConfig
		imageName string
		tag       string
		want      string
	}{
		{
			name: "Docker Hub",
			registry: &config.RegistryConfig{
				Server:   "docker.io",
				Username: "myuser",
			},
			imageName: "myapp",
			tag:       "v1",
			want:      "myuser/myapp:v1",
		},
		{
			name: "GitHub Container Registry",
			registry: &config.RegistryConfig{
				Server:   "ghcr.io",
				Username: "myuser",
			},
			imageName: "myapp",
			tag:       "abc123",
			want:      "ghcr.io/myuser/myapp:abc123",
		},
		{
			name: "GitLab",
			registry: &config.RegistryConfig{
				Server:   "registry.gitlab.com",
				Username: "myuser",
			},
			imageName: "myapp",
			tag:       "latest",
			want:      "registry.gitlab.com/myuser/myapp:latest",
		},
		{
			name: "Private Registry",
			registry: &config.RegistryConfig{
				Server:   "registry.example.com",
				Username: "admin",
			},
			imageName: "myapp",
			tag:       "v2.0",
			want:      "registry.example.com/admin/myapp:v2.0",
		},
		{
			name: "No registry",
			registry: nil,
			imageName: "myapp",
			tag:       "v1",
			want:      "myapp:v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{config: tt.registry}
			got := client.GetImagePath(tt.imageName, tt.tag)
			if got != tt.want {
				t.Errorf("GetImagePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConfigured(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
		want   bool
	}{
		{
			name: "fully configured",
			config: &config.Config{
				Registry: &config.RegistryConfig{
					Server:   "ghcr.io",
					Username: "user",
					Password: "pass",
				},
			},
			want: true,
		},
		{
			name: "no registry",
			config: &config.Config{
				Registry: nil,
			},
			want: false,
		},
		{
			name: "missing username",
			config: &config.Config{
				Registry: &config.RegistryConfig{
					Server:   "ghcr.io",
					Password: "pass",
				},
			},
			want: false,
		},
		{
			name: "missing password",
			config: &config.Config{
				Registry: &config.RegistryConfig{
					Server:   "ghcr.io",
					Username: "user",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConfigured(tt.config)
			if got != tt.want {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPull_RequiresConfig(t *testing.T) {
	client := &Client{config: nil}
	mockSSH := ssh.NewMockClient()
	
	err := client.Pull(mockSSH, "myapp", "v1")
	if err == nil {
		t.Error("Pull() should fail without registry config")
	}
}

func TestLogin_RequiresConfig(t *testing.T) {
	client := &Client{config: nil}
	
	err := client.Login()
	if err == nil {
		t.Error("Login() should fail without registry config")
	}
}

