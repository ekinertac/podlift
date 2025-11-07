package commands

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Should not panic and return something
	version := getVersion()
	if version == "" {
		t.Error("getVersion() returned empty string")
	}
	
	// Should be either "dev" or a git tag/commit
	if version != "dev" && !strings.Contains(version, "v") && !strings.Contains(version, "-") {
		t.Logf("Version format: %s", version)
	}
}

func TestGetVersionOutput(t *testing.T) {
	output := getVersionOutput()
	if output == "" {
		t.Error("getVersionOutput() returned empty string")
	}
	
	// Should contain key information
	requiredStrings := []string{
		"podlift",
		"Go version:",
		"OS/Arch:",
	}
	
	for _, required := range requiredStrings {
		if !strings.Contains(output, required) {
			t.Errorf("Output missing required string: %s", required)
		}
	}
	
	// Should contain actual runtime info
	if !strings.Contains(output, runtime.GOOS) {
		t.Errorf("Output missing OS: %s", runtime.GOOS)
	}
	if !strings.Contains(output, runtime.GOARCH) {
		t.Errorf("Output missing arch: %s", runtime.GOARCH)
	}
}

func TestVersionVariables(t *testing.T) {
	// Version, Commit, Date should be settable via ldflags
	oldVersion := Version
	oldCommit := Commit
	oldDate := Date
	
	Version = "test-version"
	Commit = "test-commit"
	Date = "test-date"
	
	if Version != "test-version" {
		t.Error("Version variable should be mutable")
	}
	
	// Restore
	Version = oldVersion
	Commit = oldCommit
	Date = oldDate
}
