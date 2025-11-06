package deploy

import (
	"testing"

	"github.com/ekinertac/podlift/internal/config"
)

// Test deployToServer with mock SSH client
func TestDeployToServer_BasicFlow(t *testing.T) {
	// Skip - requires real Docker build
	// This is tested in E2E
	t.Skip("Requires Docker - E2E tested")
}

func TestDeployOptions_Values(t *testing.T) {
	opts := DeployOptions{
		Config:     &config.Config{Service: "test"},
		SkipBuild:  true,
		SkipHealth: true,
		Parallel:   false,
		DryRun:     true,
	}

	if opts.Config.Service != "test" {
		t.Error("Config should be set")
	}
	if !opts.SkipBuild {
		t.Error("SkipBuild should be true")
	}
	if !opts.SkipHealth {
		t.Error("SkipHealth should be true")
	}
	if opts.Parallel {
		t.Error("Parallel should be false")
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
}

// Note: Full deployment requires:
// - Real Docker daemon (for build)
// - Real git repository  
// - Real file system (for tar creation)
// These are all tested in E2E tests with Multipass

// What we CAN test: validation and error handling
func TestDeploy_ValidatesGitState(t *testing.T) {
	t.Skip("Requires git repository - tested in E2E")
}

