package ssh

import (
	"strings"
	"testing"
)

func TestNormalizeGitURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "trailing .git",
			input: "https://github.com/user/repo.git",
			want:  "user/repo",
		},
		{
			name:  "no .git",
			input: "https://github.com/user/repo",
			want:  "user/repo",
		},
		{
			name:  "ssh format",
			input: "git@github.com:user/repo.git",
			want:  "user/repo",
		},
		{
			name:  "uppercase",
			input: "https://github.com/User/Repo",
			want:  "user/repo",
		},
		{
			name:  "with spaces",
			input: "  https://github.com/user/repo.git  ",
			want:  "user/repo",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeGitURL(tt.input)
			if got != tt.want {
				t.Errorf("normalizeGitURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestServiceConflictError_Message(t *testing.T) {
	err := &ServiceConflictError{
		ServiceName: "test-service",
		Info: &ServiceInfo{
			Name:       "test-service",
			Containers: []string{"test-service-web-1", "test-service-web-2"},
			DeployedAt: "2025-11-05T10:30:00Z",
			Version:    "abc123",
		},
	}

	msg := err.Error()

	// Check all required elements are in message
	checks := []string{
		"test-service",
		"abc123",
		"Containers: 2",
		"same application",
		"DIFFERENT application",
		"Change the service name",
	}

	for _, check := range checks {
		if !strings.Contains(msg, check) {
			t.Errorf("Error message missing %q\nMessage: %s", check, msg)
		}
	}
}

func TestServiceInfo_Structure(t *testing.T) {
	info := &ServiceInfo{
		Name:       "myapp",
		Containers: []string{"myapp-web-1"},
		DeployedAt: "2025-11-05T10:30:00Z",
		Version:    "abc123",
	}

	if info.Name != "myapp" {
		t.Errorf("Name = %v, want myapp", info.Name)
	}
	if len(info.Containers) != 1 {
		t.Errorf("Containers length = %d, want 1", len(info.Containers))
	}
	if info.Version != "abc123" {
		t.Errorf("Version = %v, want abc123", info.Version)
	}
}

