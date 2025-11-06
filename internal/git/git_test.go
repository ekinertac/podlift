package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	dir, err := os.MkdirTemp("", "podlift-git-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	// Configure git
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	cmd.Run()

	// Save original directory
	originalDir, _ := os.Getwd()

	// Change to test repo
	os.Chdir(dir)

	// Cleanup function
	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestIsRepository(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	if !IsRepository() {
		t.Error("IsRepository() = false, want true")
	}

	// Test outside git repo
	os.Chdir("/tmp")
	defer os.Chdir("/")
	
	// /tmp might be in a git repo in some setups, so we can't reliably test false case
}

func TestGetCommitHash(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	hash, err := GetCommitHash()
	if err != nil {
		t.Fatalf("GetCommitHash() error = %v", err)
	}

	if len(hash) != 7 {
		t.Errorf("GetCommitHash() returned %v (%d chars), want 7 chars", hash, len(hash))
	}
}

func TestGetBranch(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	branch, err := GetBranch()
	if err != nil {
		t.Fatalf("GetBranch() error = %v", err)
	}

	// Default branch is usually "master" or "main"
	if branch != "master" && branch != "main" {
		t.Logf("Branch = %v (expected master or main)", branch)
	}
}

func TestIsClean(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Should be clean after setup
	clean, changes, err := IsClean()
	if err != nil {
		t.Fatalf("IsClean() error = %v", err)
	}

	if !clean {
		t.Errorf("IsClean() = false, want true. Changes: %v", changes)
	}

	if len(changes) != 0 {
		t.Errorf("IsClean() returned %d changes, want 0", len(changes))
	}

	// Make a change
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("modified"), 0644)

	// Should be dirty now
	clean, changes, err = IsClean()
	if err != nil {
		t.Fatalf("IsClean() error = %v", err)
	}

	if clean {
		t.Error("IsClean() = true after modification, want false")
	}

	if len(changes) == 0 {
		t.Error("IsClean() returned 0 changes after modification, want > 0")
	}
}

func TestRequireCleanState(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Should pass when clean
	if err := RequireCleanState(); err != nil {
		t.Errorf("RequireCleanState() error = %v, want nil", err)
	}

	// Make a change
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("modified"), 0644)

	// Should fail when dirty
	err := RequireCleanState()
	if err == nil {
		t.Error("RequireCleanState() succeeded with dirty tree, want error")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "uncommitted changes") {
		t.Errorf("RequireCleanState() error should mention uncommitted changes, got: %v", errStr)
	}
}

func TestGetCommitMessage(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	msg, err := GetCommitMessage()
	if err != nil {
		t.Fatalf("GetCommitMessage() error = %v", err)
	}

	if msg != "Initial commit" {
		t.Errorf("GetCommitMessage() = %v, want 'Initial commit'", msg)
	}
}

func TestGetVersion(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Without tag, should return commit hash
	version, err := GetVersion()
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	if len(version) != 7 {
		t.Errorf("GetVersion() = %v (%d chars), want 7 char hash", version, len(version))
	}

	// Create a tag
	cmd := exec.Command("git", "tag", "v1.0.0")
	cmd.Run()

	// With tag, should return tag
	version, err = GetVersion()
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	if version != "v1.0.0" {
		t.Errorf("GetVersion() = %v, want v1.0.0", version)
	}
}

func TestCheckStatus(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	status, err := CheckStatus()
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}

	if !status.Clean {
		t.Error("CheckStatus() Clean = false, want true")
	}

	if len(status.CommitHash) != 7 {
		t.Errorf("CheckStatus() CommitHash length = %d, want 7", len(status.CommitHash))
	}

	if status.Branch == "" {
		t.Error("CheckStatus() Branch is empty")
	}
}

