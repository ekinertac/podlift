package commands

import (
	"runtime"
	"strings"
	"testing"
)

func TestOsName(t *testing.T) {
	name := osName()
	if name == "" {
		t.Error("osName() should not return empty string")
	}
	
	// Should be one of the known OS names
	validOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
		"freebsd": true,
	}
	
	if !validOS[name] {
		t.Logf("Unexpected OS: %s (but might be valid)", name)
	}
}

func TestOsArch(t *testing.T) {
	arch := osArch()
	if arch == "" {
		t.Error("osArch() should not return empty string")
	}
	
	// Should match runtime
	if arch != runtime.GOARCH {
		t.Errorf("osArch() = %v, want %v", arch, runtime.GOARCH)
	}
}

func TestGoVersion(t *testing.T) {
	version := goVersion()
	if version == "" {
		t.Error("goVersion() should not return empty string")
	}
	
	// Should contain "go"
	if !strings.Contains(version, "go") {
		t.Errorf("goVersion() should contain 'go', got: %v", version)
	}
}

func TestVersionVariables(t *testing.T) {
	// Version, Commit, Date should be settable
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

