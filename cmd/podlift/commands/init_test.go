package commands

import (
	"os"
	"strings"
	"testing"
)

func TestDetectProjectName(t *testing.T) {
	tests := []struct {
		dir  string
		want string
	}{
		{"/home/user/myapp", "myapp"},
		{"/home/user/My_App", "my-app"},
		{"/home/user/MY-APP", "my-app"},
		{"/home/user/app_name_123", "app-name-123"},
		{"/home/user/@#$%", "myapp"}, // Invalid chars -> default
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			got := detectProjectName(tt.dir)
			if got != tt.want {
				t.Errorf("detectProjectName(%v) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create temp file
	tmpfile, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	if !fileExists(tmpfile.Name()) {
		t.Error("fileExists() should return true for existing file")
	}

	if fileExists("/nonexistent/file/path") {
		t.Error("fileExists() should return false for non-existent file")
	}
}

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		hasDockerfile bool
		checks        []string
	}{
		{
			name:          "basic config",
			serviceName:   "myapp",
			hasDockerfile: true,
			checks: []string{
				"service: myapp",
				"image: myapp",
				"servers:",
				"YOUR_SERVER_IP",
			},
		},
		{
			name:          "no dockerfile warning",
			serviceName:   "myapp",
			hasDockerfile: false,
			checks: []string{
				"# WARNING: No Dockerfile found",
				"service: myapp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateConfig(tt.serviceName, tt.hasDockerfile)

			for _, check := range tt.checks {
				if !strings.Contains(got, check) {
					t.Errorf("generateConfig() missing %q", check)
				}
			}
		})
	}
}

func TestGenerateEnvExample(t *testing.T) {
	got := generateEnvExample()

	checks := []string{
		"REGISTRY_USER",
		"REGISTRY_PASSWORD",
		"SECRET_KEY",
		"Never commit .env to git",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("generateEnvExample() missing %q", check)
		}
	}
}

func TestRunInit_Integration(t *testing.T) {
	// Create temp directory
	tmpdir, err := os.MkdirTemp("", "test-init-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(oldDir)

	// Run init
	cmd := initCommand
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	// Check files were created
	if _, err := os.Stat("podlift.yml"); os.IsNotExist(err) {
		t.Error("podlift.yml was not created")
	}

	if _, err := os.Stat(".env.example"); os.IsNotExist(err) {
		t.Error(".env.example was not created")
	}

	// Run init again - should fail
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("runInit() should fail when podlift.yml exists")
	}
}

