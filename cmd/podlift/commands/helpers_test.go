package commands

import (
	"testing"
)

func TestTrimVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "Docker version 24.0.5, build abc123",
			want:  "Docker version 24.0.5, build abc123",
		},
		{
			input: "Docker version 24.0.5\nExtra line",
			want:  "Docker version 24.0.5",
		},
		{
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := trimVersion(tt.input)
			if got != tt.want {
				t.Errorf("trimVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"line1\nline2\nline3", 3},
		{"line1\n\nline2", 2}, // Empty lines removed
		{"  line1  \n  line2  ", 2}, // Trimmed
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != tt.want {
				t.Errorf("splitLines() returned %d lines, want %d", len(got), tt.want)
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
		{"myapp-worker-def456-2", "worker"},
		{"simple", "simple"},
		{"a-b-c-d", "b"},
	}

	for _, tt := range tests {
		t.Run(tt.container, func(t *testing.T) {
			got := extractServiceName(tt.container)
			if got != tt.want {
				t.Errorf("extractServiceName(%v) = %v, want %v", tt.container, got, tt.want)
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
		{"Exited (0) 5 minutes ago", "-"},
		{"Created", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := extractUptime(tt.status)
			if got != tt.want {
				t.Errorf("extractUptime(%v) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

