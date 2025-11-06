package ssh

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// Note: Most SCP tests require real SSH connection
// These are tested in integration tests

func TestCopyFile_RequiresExistingFile(t *testing.T) {
	// This would test that file must exist
	// But we can't test without real connection
	// Skip for now - covered in E2E tests
	t.Skip("Requires real SSH connection - tested in E2E")
}

// Integration test for actual file copy
func TestCopyFile_Integration(t *testing.T) {
	host := getTestHost()
	if host == "" {
		t.Skip("Skipping integration test: set PODLIFT_SSH_TEST_HOST to enable")
	}

	// Create test file
	tmpfile, err := os.CreateTemp("", "scp-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "test file content"
	tmpfile.WriteString(content)
	tmpfile.Close()

	// Create SSH client
	client, err := NewClient(Config{
		Host:    host,
		User:    "root",
		KeyPath: "~/.ssh/id_rsa",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	// Copy file
	remotePath := "/tmp/scp-test.txt"
	err = client.CopyFile(tmpfile.Name(), remotePath)
	if err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	// Verify file exists on remote
	output, err := client.Execute(fmt.Sprintf("cat %s", remotePath))
	if err != nil {
		t.Fatalf("Failed to read remote file: %v", err)
	}

	if !strings.Contains(output, content) {
		t.Errorf("Remote file content = %v, want %v", output, content)
	}

	// Cleanup
	client.Execute(fmt.Sprintf("rm %s", remotePath))
}

func TestCopyFileWithProgress_Integration(t *testing.T) {
	host := getTestHost()
	if host == "" {
		t.Skip("Skipping integration test: set PODLIFT_SSH_TEST_HOST to enable")
	}

	// Create test file
	tmpfile, err := os.CreateTemp("", "scp-progress-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some data
	data := make([]byte, 1024*100) // 100KB
	tmpfile.Write(data)
	tmpfile.Close()

	client, err := NewClient(Config{
		Host:    host,
		User:    "root",
		KeyPath: "~/.ssh/id_rsa",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	// Track progress
	progressCalled := false
	remotePath := "/tmp/scp-progress-test.txt"

	err = client.CopyFileWithProgress(tmpfile.Name(), remotePath, func(sent, total int64) {
		progressCalled = true
		if sent < 0 || total < 0 || sent > total {
			t.Errorf("Invalid progress values: sent=%d, total=%d", sent, total)
		}
	})

	if err != nil {
		t.Fatalf("CopyFileWithProgress() error = %v", err)
	}

	if !progressCalled {
		t.Error("Progress callback was not called")
	}

	// Cleanup
	client.Execute(fmt.Sprintf("rm %s", remotePath))
}

