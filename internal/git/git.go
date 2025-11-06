package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Status represents git repository status
type Status struct {
	Clean      bool
	CommitHash string
	Branch     string
	Tag        string
	Changes    []string
}

// CheckStatus checks the current git repository status
func CheckStatus() (*Status, error) {
	// Check if we're in a git repository
	if !IsRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	status := &Status{}

	// Get commit hash
	hash, err := GetCommitHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}
	status.CommitHash = hash

	// Get branch name
	branch, err := GetBranch()
	if err != nil {
		// Not an error if detached HEAD
		status.Branch = ""
	} else {
		status.Branch = branch
	}

	// Get tag if on tagged commit
	tag, _ := GetTag()
	status.Tag = tag

	// Check if working tree is clean
	clean, changes, err := IsClean()
	if err != nil {
		return nil, fmt.Errorf("failed to check git status: %w", err)
	}
	status.Clean = clean
	status.Changes = changes

	return status, nil
}

// IsRepository checks if current directory is a git repository
func IsRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetCommitHash returns the short commit hash (7 chars)
func GetCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitHashLong returns the full commit hash
func GetCommitHashLong() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetBranch returns the current branch name
func GetBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		return "", fmt.Errorf("detached HEAD state")
	}
	return branch, nil
}

// GetTag returns the tag name if current commit is tagged
func GetTag() (string, error) {
	cmd := exec.Command("git", "describe", "--exact-match", "--tags", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err // Not tagged
	}
	return strings.TrimSpace(string(output)), nil
}

// IsClean checks if working tree has uncommitted changes
func IsClean() (bool, []string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, nil, err
	}

	if len(output) == 0 {
		return true, nil, nil
	}

	// Parse changed files
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var changes []string
	for _, line := range lines {
		if line != "" {
			changes = append(changes, line)
		}
	}

	return false, changes, nil
}

// GetCommitMessage returns the commit message for the current HEAD
func GetCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL returns the remote URL (origin)
func GetRemoteURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetVersion returns a version string combining tag and commit
// Returns tag if on tagged commit, otherwise commit hash
func GetVersion() (string, error) {
	tag, err := GetTag()
	if err == nil && tag != "" {
		return tag, nil
	}

	hash, err := GetCommitHash()
	if err != nil {
		return "", err
	}
	return hash, nil
}

// RequireCleanState returns error if working tree is dirty
func RequireCleanState() error {
	clean, changes, err := IsClean()
	if err != nil {
		return err
	}

	if !clean {
		var buf bytes.Buffer
		buf.WriteString("working tree has uncommitted changes\n\n")
		buf.WriteString("Modified files:\n")
		for _, change := range changes {
			buf.WriteString("  " + change + "\n")
		}
		buf.WriteString("\nCommit or stash your changes before deploying.")
		return fmt.Errorf("%s", buf.String())
	}

	return nil
}

// GetInfo returns a formatted info string about current git state
func GetInfo() (string, error) {
	status, err := CheckStatus()
	if err != nil {
		return "", err
	}

	var info []string
	
	info = append(info, fmt.Sprintf("Commit: %s", status.CommitHash))
	
	if status.Branch != "" {
		info = append(info, fmt.Sprintf("Branch: %s", status.Branch))
	}
	
	if status.Tag != "" {
		info = append(info, fmt.Sprintf("Tag: %s", status.Tag))
	}

	msg, err := GetCommitMessage()
	if err == nil {
		// Truncate long messages
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}
		info = append(info, fmt.Sprintf("Message: %s", msg))
	}

	if !status.Clean {
		info = append(info, fmt.Sprintf("Status: DIRTY (%d changes)", len(status.Changes)))
	} else {
		info = append(info, "Status: Clean")
	}

	return strings.Join(info, "\n"), nil
}

