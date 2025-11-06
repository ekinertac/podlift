package docker

import (
	"strings"
	"testing"
)

func TestGenerateRunCommand(t *testing.T) {
	tests := []struct {
		name   string
		config ContainerConfig
		checks []string // Strings that should be in output
	}{
		{
			name: "basic container",
			config: ContainerConfig{
				Name:  "myapp-web-1",
				Image: "myapp:abc123",
				Port:  8000,
			},
			checks: []string{
				"docker run -d",
				"--name myapp-web-1",
				"-p 8000:8000",
				"myapp:abc123",
			},
		},
		{
			name: "with environment variables",
			config: ContainerConfig{
				Name:  "myapp-web-1",
				Image: "myapp:abc123",
				Env: map[string]string{
					"SECRET_KEY": "test123",
					"DEBUG":      "false",
				},
			},
			checks: []string{
				"-e SECRET_KEY=",
				"-e DEBUG=",
			},
		},
		{
			name: "with labels",
			config: ContainerConfig{
				Name:  "myapp-web-1",
				Image: "myapp:abc123",
				Labels: map[string]string{
					"podlift.service": "myapp",
					"podlift.version": "abc123",
				},
			},
			checks: []string{
				"--label podlift.service=myapp",
				"--label podlift.version=abc123",
			},
		},
		{
			name: "with volumes",
			config: ContainerConfig{
				Name:    "myapp-web-1",
				Image:   "myapp:abc123",
				Volumes: []string{"/data:/app/data", "uploads:/app/uploads"},
			},
			checks: []string{
				"-v /data:/app/data",
				"-v uploads:/app/uploads",
			},
		},
		{
			name: "with custom command",
			config: ContainerConfig{
				Name:    "myapp-worker-1",
				Image:   "myapp:abc123",
				Command: "celery worker",
			},
			checks: []string{
				"myapp:abc123",
				"celery worker",
			},
		},
		{
			name: "with port mapping",
			config: ContainerConfig{
				Name:         "myapp-web-1",
				Image:        "myapp:abc123",
				Port:         9000,
				InternalPort: 8000,
			},
			checks: []string{
				"-p 9000:8000",
			},
		},
		{
			name: "with options",
			config: ContainerConfig{
				Name:  "myapp-web-1",
				Image: "myapp:abc123",
				Options: map[string]string{
					"memory": "512m",
					"cpus":   "1",
					"restart": "always",
				},
			},
			checks: []string{
				"--memory=512m",
				"--cpus=1",
				"--restart=always",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateRunCommand(tt.config)

			for _, check := range tt.checks {
				if !strings.Contains(got, check) {
					t.Errorf("GenerateRunCommand() missing %q\nGot: %v", check, got)
				}
			}
		})
	}
}

func TestExtractUptime(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"Up 5 minutes", "5m"},
		{"Up 2 hours", "2h"},
		{"Up 3 days", "3d"},
		{"Up 30 seconds", "<1m"},
		{"Exited (0) 1 hour ago", "-"},
		{"", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			// This function is in ps.go, but testing the pattern
			// We'll import it or duplicate for testing
			got := extractUptime(tt.status)
			if got != tt.want {
				t.Errorf("extractUptime(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		container string
		want      string
	}{
		{"myapp-web-abc123-1", "web"},
		{"myapp-worker-abc123-1", "worker"},
		{"simple-name", "name"},
	}

	for _, tt := range tests {
		t.Run(tt.container, func(t *testing.T) {
			got := extractServiceName(tt.container)
			if got != tt.want {
				t.Errorf("extractServiceName(%q) = %v, want %v", tt.container, got, tt.want)
			}
		})
	}
}

// Helper functions copied from ps.go for testing
func extractServiceName(containerName string) string {
	parts := strings.Split(containerName, "-")
	if len(parts) >= 2 {
		return parts[1]
	}
	return containerName
}

func extractUptime(status string) string {
	if !strings.Contains(status, "Up ") {
		return "-"
	}

	if strings.Contains(status, "second") {
		return "<1m"
	}
	if strings.Contains(status, "minute") {
		parts := strings.Split(status, " ")
		for i, part := range parts {
			if part == "Up" && i+1 < len(parts) {
				return parts[i+1] + "m"
			}
		}
	}
	if strings.Contains(status, "hour") {
		parts := strings.Split(status, " ")
		for i, part := range parts {
			if part == "Up" && i+1 < len(parts) {
				return parts[i+1] + "h"
			}
		}
	}
	if strings.Contains(status, "day") {
		parts := strings.Split(status, " ")
		for i, part := range parts {
			if part == "Up" && i+1 < len(parts) {
				return parts[i+1] + "d"
			}
		}
	}

	return "running"
}

