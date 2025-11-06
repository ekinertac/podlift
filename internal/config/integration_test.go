package config

import (
	"os"
	"testing"
)

// TestLoadMinimalConfig tests loading the minimal test config
func TestLoadMinimalConfig(t *testing.T) {
	config, err := Load("../../testdata/minimal.yml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Service != "myapp" {
		t.Errorf("Service = %v, want myapp", config.Service)
	}

	if config.Image != "myapp" {
		t.Errorf("Image = %v, want myapp", config.Image)
	}

	servers := config.Servers.Get()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server role, got %d", len(servers))
	}

	webServers := servers["web"]
	if len(webServers) != 1 {
		t.Errorf("Expected 1 web server, got %d", len(webServers))
	}

	if webServers[0].Host != "192.168.1.10" {
		t.Errorf("Server host = %v, want 192.168.1.10", webServers[0].Host)
	}

	// Check defaults were applied
	if webServers[0].User != "root" {
		t.Errorf("Server user = %v, want root (default)", webServers[0].User)
	}

	if webServers[0].Port != 22 {
		t.Errorf("Server port = %v, want 22 (default)", webServers[0].Port)
	}
}

// TestLoadStandardConfig tests loading the standard test config
func TestLoadStandardConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("REGISTRY_USER", "testuser")
	os.Setenv("REGISTRY_PASSWORD", "testpass")
	os.Setenv("DB_PASSWORD", "dbpass")
	os.Setenv("SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("REGISTRY_USER")
		os.Unsetenv("REGISTRY_PASSWORD")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("SECRET_KEY")
	}()

	config, err := Load("../../testdata/standard.yml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check basic fields
	if config.Service != "myapp" {
		t.Errorf("Service = %v, want myapp", config.Service)
	}

	if config.Domain != "myapp.com" {
		t.Errorf("Domain = %v, want myapp.com", config.Domain)
	}

	// Should have one server
	servers := config.Servers.Get()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server role, got %d", len(servers))
	}

	// Should have postgres dependency
	if len(config.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(config.Dependencies))
	}

	postgres, ok := config.Dependencies["postgres"]
	if !ok {
		t.Fatal("postgres dependency not found")
	}
	if postgres.Image != "postgres:16" {
		t.Errorf("postgres image = %v, want postgres:16", postgres.Image)
	}

	// Should have web service with 2 replicas
	web, ok := config.Services["web"]
	if !ok {
		t.Fatal("web service not found")
	}
	if web.Replicas != 2 {
		t.Errorf("web replicas = %v, want 2", web.Replicas)
	}

	// Should have SSL enabled
	if config.Proxy == nil || config.Proxy.SSL != "letsencrypt" {
		t.Error("SSL should be configured with letsencrypt")
	}

	// Should have after_deploy hook
	if config.Hooks == nil || len(config.Hooks.AfterDeploy) == 0 {
		t.Error("Should have after_deploy hooks")
	}
}

// TestLoadFullConfig tests loading the full test config
func TestLoadFullConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("REGISTRY_USER", "testuser")
	os.Setenv("REGISTRY_PASSWORD", "testpass")
	os.Setenv("DB_PASSWORD", "dbpass")
	os.Setenv("SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("REGISTRY_USER")
		os.Unsetenv("REGISTRY_PASSWORD")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("SECRET_KEY")
	}()

	config, err := Load("../../testdata/full.yml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check basic fields
	if config.Service != "myapp" {
		t.Errorf("Service = %v, want myapp", config.Service)
	}

	if config.Domain != "myapp.com" {
		t.Errorf("Domain = %v, want myapp.com", config.Domain)
	}

	// Check servers
	servers := config.Servers.Get()
	if len(servers) != 2 {
		t.Errorf("Expected 2 server roles, got %d", len(servers))
	}

	webServers := servers["web"]
	if len(webServers) != 2 {
		t.Errorf("Expected 2 web servers, got %d", len(webServers))
	}

	workerServers := servers["worker"]
	if len(workerServers) != 1 {
		t.Errorf("Expected 1 worker server, got %d", len(workerServers))
	}

	// Check primary server
	primary, role, err := config.GetPrimaryServer()
	if err != nil {
		t.Errorf("GetPrimaryServer() error = %v", err)
	}
	if primary.Host != "192.168.1.10" {
		t.Errorf("Primary server = %v, want 192.168.1.10", primary.Host)
	}
	if role != "web" {
		t.Errorf("Primary role = %v, want web", role)
	}

	// Check registry
	if config.Registry == nil {
		t.Fatal("Registry is nil")
	}
	if config.Registry.Server != "ghcr.io" {
		t.Errorf("Registry server = %v, want ghcr.io", config.Registry.Server)
	}

	// Environment variables should be substituted
	config.SubstituteConfigEnvVars()
	if config.Registry.Username != "testuser" {
		t.Errorf("Registry username = %v, want testuser", config.Registry.Username)
	}

	// Check dependencies
	if len(config.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(config.Dependencies))
	}

	postgres, ok := config.Dependencies["postgres"]
	if !ok {
		t.Fatal("postgres dependency not found")
	}
	if postgres.Image != "postgres:16" {
		t.Errorf("postgres image = %v, want postgres:16", postgres.Image)
	}
	if postgres.Port != 5432 {
		t.Errorf("postgres port = %v, want 5432", postgres.Port)
	}

	redis, ok := config.Dependencies["redis"]
	if !ok {
		t.Fatal("redis dependency not found")
	}
	if redis.Image != "redis:7-alpine" {
		t.Errorf("redis image = %v, want redis:7-alpine", redis.Image)
	}

	// Check services
	if len(config.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(config.Services))
	}

	web, ok := config.Services["web"]
	if !ok {
		t.Fatal("web service not found")
	}
	if web.Port != 8000 {
		t.Errorf("web port = %v, want 8000", web.Port)
	}
	if web.Replicas != 2 {
		t.Errorf("web replicas = %v, want 2", web.Replicas)
	}

	worker, ok := config.Services["worker"]
	if !ok {
		t.Fatal("worker service not found")
	}
	if worker.Command != "celery -A myapp worker" {
		t.Errorf("worker command = %v, want 'celery -A myapp worker'", worker.Command)
	}

	// Check proxy
	if config.Proxy == nil {
		t.Fatal("Proxy is nil")
	}
	if !config.Proxy.Enabled {
		t.Error("Proxy should be enabled")
	}
	if config.Proxy.SSL != "letsencrypt" {
		t.Errorf("Proxy SSL = %v, want letsencrypt", config.Proxy.SSL)
	}

	// Check hooks
	if config.Hooks == nil {
		t.Fatal("Hooks is nil")
	}
	if len(config.Hooks.AfterDeploy) != 2 {
		t.Errorf("Expected 2 after_deploy hooks, got %d", len(config.Hooks.AfterDeploy))
	}
}

// TestGetAllServers tests getting all servers
func TestGetAllServers(t *testing.T) {
	config, err := Load("../../testdata/full.yml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	allServers := config.GetAllServers()
	if len(allServers) != 3 {
		t.Errorf("Expected 3 total servers, got %d", len(allServers))
	}

	// Check roles are correct
	webCount := 0
	workerCount := 0
	for _, s := range allServers {
		switch s.Role {
		case "web":
			webCount++
		case "worker":
			workerCount++
		}
	}

	if webCount != 2 {
		t.Errorf("Expected 2 web servers, got %d", webCount)
	}
	if workerCount != 1 {
		t.Errorf("Expected 1 worker server, got %d", workerCount)
	}
}

// TestEnvVarSubstitution tests environment variable substitution
func TestEnvVarSubstitution(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envKey   string
		envValue string
		want     string
	}{
		{
			name:     "simple substitution",
			input:    "${TEST_VAR}",
			envKey:   "TEST_VAR",
			envValue: "test_value",
			want:     "test_value",
		},
		{
			name:     "with default",
			input:    "${MISSING_VAR:-default}",
			envKey:   "",
			envValue: "",
			want:     "default",
		},
		{
			name:     "in string",
			input:    "prefix_${TEST_VAR}_suffix",
			envKey:   "TEST_VAR",
			envValue: "middle",
			want:     "prefix_middle_suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := SubstituteEnvVars(tt.input)
			if result != tt.want {
				t.Errorf("SubstituteEnvVars(%v) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}
