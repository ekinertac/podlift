package ssh

import (
	"fmt"
	"strings"
	"testing"
)

func TestServiceConflictError(t *testing.T) {
	err := &ServiceConflictError{
		ServiceName: "myapp",
		Info: &ServiceInfo{
			Name:       "myapp",
			Version:    "abc123",
			DeployedAt: "2025-11-05T10:30:00Z",
			Containers: []string{"myapp-web-abc123-1", "myapp-web-abc123-2"},
		},
	}

	errMsg := err.Error()
	
	// Should contain service name
	if !strings.Contains(errMsg, "myapp") {
		t.Error("Error message should contain service name")
	}

	// Should contain version
	if !strings.Contains(errMsg, "abc123") {
		t.Error("Error message should contain version")
	}

	// Should contain helpful messages
	if !strings.Contains(errMsg, "same application") {
		t.Error("Error message should mention same application scenario")
	}
	if !strings.Contains(errMsg, "DIFFERENT application") {
		t.Error("Error message should mention different application scenario")
	}
}

func TestIsConflictError(t *testing.T) {
	conflictErr := &ServiceConflictError{
		ServiceName: "test",
		Info:        &ServiceInfo{},
	}

	if !IsConflictError(conflictErr) {
		t.Error("IsConflictError should return true for ServiceConflictError")
	}

	normalErr := fmt.Errorf("normal error")
	if IsConflictError(normalErr) {
		t.Error("IsConflictError should return false for normal error")
	}
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		url1 string
		url2 string
		same bool
	}{
		{
			url1: "git@github.com:user/repo.git",
			url2: "https://github.com/user/repo",
			same: true,
		},
		{
			url1: "git@github.com:user/repo",
			url2: "https://github.com/user/repo.git",
			same: true,
		},
		{
			url1: "git@github.com:user/repo1.git",
			url2: "git@github.com:user/repo2.git",
			same: false,
		},
		{
			url1: "https://github.com/User/Repo.git",
			url2: "https://github.com/user/repo",
			same: true, // Case insensitive
		},
	}

	for _, tt := range tests {
		norm1 := normalizeGitURL(tt.url1)
		norm2 := normalizeGitURL(tt.url2)
		
		if (norm1 == norm2) != tt.same {
			t.Errorf("normalizeGitURL(%v, %v) same=%v, want %v", 
				tt.url1, tt.url2, norm1 == norm2, tt.same)
		}
	}
}

// Integration tests for existing service detection
// These require a real server with Docker

func TestCheckExistingService_Integration(t *testing.T) {
	host := getTestHost()
	if host == "" {
		t.Skip("Skipping integration test: set PODLIFT_SSH_TEST_HOST to enable")
	}

	client, err := NewClient(Config{
		Host:    host,
		User:    "root",
		KeyPath: "~/.ssh/id_rsa",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	// Check for non-existent service
	exists, info, err := client.CheckExistingService("nonexistent-service-xyz")
	if err != nil {
		t.Fatalf("CheckExistingService() error = %v", err)
	}

	if exists {
		t.Error("CheckExistingService() should return false for non-existent service")
	}

	if info != nil {
		t.Error("CheckExistingService() should return nil info for non-existent service")
	}
}

func TestListPodliftServices_Integration(t *testing.T) {
	host := getTestHost()
	if host == "" {
		t.Skip("Skipping integration test: set PODLIFT_SSH_TEST_HOST to enable")
	}

	client, err := NewClient(Config{
		Host:    host,
		User:    "root",
		KeyPath: "~/.ssh/id_rsa",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	services, err := client.ListPodliftServices()
	if err != nil {
		t.Fatalf("ListPodliftServices() error = %v", err)
	}

	// Should return a list (might be empty)
	if services == nil {
		t.Error("ListPodliftServices() should return empty slice, not nil")
	}
}

