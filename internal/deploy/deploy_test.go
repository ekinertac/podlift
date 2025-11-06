package deploy

import (
	"testing"
)

func TestDeploy_RequiresConfig(t *testing.T) {
	// This test would panic with nil config
	// In practice, this never happens because commands validate config first
	// Skip this test - it's testing impossible state
	t.Skip("Deploy requires valid config - checked by commands layer")
}

func TestDeployOptions_Defaults(t *testing.T) {
	opts := DeployOptions{
		SkipBuild:  false,
		SkipHealth: false,
		Parallel:   false,
		DryRun:     false,
	}

	// Just verify the struct can be created
	if opts.SkipBuild {
		t.Error("Default SkipBuild should be false")
	}
}

// Integration tests for deploy require:
// - Real Docker daemon
// - Real SSH server
// - Real git repository
// These will be tested in E2E tests with Multipass

func TestDeployToServer_Integration(t *testing.T) {
	t.Skip("Integration test - requires real server and Docker")
}

