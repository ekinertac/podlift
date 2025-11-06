package deploy

import (
	"testing"
)

// Most deploy functions require real infrastructure
// Test what we can in isolation

func TestDeployOptions_Structure(t *testing.T) {
	opts := DeployOptions{
		SkipBuild:  true,
		SkipHealth: false,
		Parallel:   true,
		DryRun:     false,
	}

	if !opts.SkipBuild {
		t.Error("SkipBuild should be true")
	}
	if opts.SkipHealth {
		t.Error("SkipHealth should be false")
	}
	if !opts.Parallel {
		t.Error("Parallel should be true")
	}
	if opts.DryRun {
		t.Error("DryRun should be false")
	}
}

// Deploy and deployToServer require:
// - Real Docker daemon
// - Real SSH connection
// - Real server
// These are tested in E2E tests

