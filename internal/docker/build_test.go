package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateLoadCommand(t *testing.T) {
	cmd := GenerateLoadCommand("/tmp/image.tar")
	expected := "sudo docker load -i /tmp/image.tar"
	if cmd != expected {
		t.Errorf("GenerateLoadCommand() = %v, want %v", cmd, expected)
	}
}

func TestGenerateStopCommand(t *testing.T) {
	cmd := GenerateStopCommand("myapp-web-1")
	expected := "docker stop myapp-web-1"
	if cmd != expected {
		t.Errorf("GenerateStopCommand() = %v, want %v", cmd, expected)
	}
}

func TestGenerateRemoveCommand(t *testing.T) {
	cmd := GenerateRemoveCommand("myapp-web-1")
	expected := "docker rm myapp-web-1"
	if cmd != expected {
		t.Errorf("GenerateRemoveCommand() = %v, want %v", cmd, expected)
	}
}

func TestGenerateLogsCommand(t *testing.T) {
	tests := []struct {
		name      string
		container string
		tail      int
		follow    bool
		want      string
	}{
		{
			name:      "basic logs",
			container: "myapp-web-1",
			tail:      0,
			follow:    false,
			want:      "docker logs myapp-web-1",
		},
		{
			name:      "with tail",
			container: "myapp-web-1",
			tail:      100,
			follow:    false,
			want:      "docker logs --tail 100 myapp-web-1",
		},
		{
			name:      "with follow",
			container: "myapp-web-1",
			tail:      0,
			follow:    true,
			want:      "docker logs -f myapp-web-1",
		},
		{
			name:      "tail and follow",
			container: "myapp-web-1",
			tail:      50,
			follow:    true,
			want:      "docker logs --tail 50 -f myapp-web-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateLogsCommand(tt.container, tt.tail, tt.follow)
			if got != tt.want {
				t.Errorf("GenerateLogsCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratePsCommand(t *testing.T) {
	cmd := GeneratePsCommand("myapp")
	
	if !strings.Contains(cmd, "docker ps") {
		t.Error("Command should contain 'docker ps'")
	}
	if !strings.Contains(cmd, "podlift.service=myapp") {
		t.Error("Command should filter by service name")
	}
}

func TestGetImageSize(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test-*.tar")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some data
	data := make([]byte, 1024*1024) // 1 MB
	tmpfile.Write(data)
	tmpfile.Close()

	size, err := GetImageSize(tmpfile.Name())
	if err != nil {
		t.Fatalf("GetImageSize() error = %v", err)
	}

	// Should be approximately 1 MB
	if size < 0.9 || size > 1.1 {
		t.Errorf("GetImageSize() = %.2f MB, want ~1.0 MB", size)
	}
}

func TestBuildImage_NoDockerfile(t *testing.T) {
	// Create temp directory without Dockerfile
	tmpdir, err := os.MkdirTemp("", "test-build-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	err = BuildImage("test", "v1", tmpdir)
	if err == nil {
		t.Error("BuildImage() should fail when Dockerfile is missing")
	}

	if !strings.Contains(err.Error(), "Dockerfile not found") {
		t.Errorf("Error should mention Dockerfile, got: %v", err)
	}
}

func TestSaveImage_CreateDirectory(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "test-save-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Try to save to nested directory that doesn't exist
	outputPath := filepath.Join(tmpdir, "nested", "dir", "image.tar")

	// This will fail because image doesn't exist, but should create directory
	SaveImage("nonexistent-image", "v1", outputPath)

	// Check if directory was created
	dir := filepath.Dir(outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("SaveImage() should create output directory")
	}
}

