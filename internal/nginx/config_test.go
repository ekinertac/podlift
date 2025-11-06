package nginx

import (
	"strings"
	"testing"
)

func TestGenerateConfig_Basic(t *testing.T) {
	cfg := Config{
		Domain:      "example.com",
		ServiceName: "myapp",
		Upstreams: []Upstream{
			{Name: "web-1", Host: "localhost", Port: 8000},
			{Name: "web-2", Host: "localhost", Port: 8001},
		},
	}

	config, err := GenerateConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v", err)
	}

	// Check required elements
	checks := []string{
		"upstream myapp",
		"server localhost:8000",
		"server localhost:8001",
		"server_name example.com",
		"proxy_pass http://myapp",
		"listen 80",
	}

	for _, check := range checks {
		if !strings.Contains(config, check) {
			t.Errorf("Config missing %q\nConfig:\n%s", check, config)
		}
	}
}

func TestGenerateConfig_WithSSL(t *testing.T) {
	cfg := Config{
		Domain:      "example.com",
		ServiceName: "myapp",
		Upstreams: []Upstream{
			{Name: "web-1", Host: "localhost", Port: 8000},
		},
		SSL: SSLConfig{
			Enabled:  true,
			CertPath: "/etc/letsencrypt/live/example.com/fullchain.pem",
			KeyPath:  "/etc/letsencrypt/live/example.com/privkey.pem",
		},
	}

	config, err := GenerateConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v", err)
	}

	checks := []string{
		"listen 443 ssl http2",
		"ssl_certificate /etc/letsencrypt",
		"return 301 https://",
		"TLSv1.2 TLSv1.3",
	}

	for _, check := range checks {
		if !strings.Contains(config, check) {
			t.Errorf("SSL config missing %q", check)
		}
	}
}

func TestGenerateConfig_CustomConf(t *testing.T) {
	cfg := Config{
		Domain:      "example.com",
		ServiceName: "myapp",
		Upstreams: []Upstream{
			{Name: "web-1", Host: "localhost", Port: 8000},
		},
		CustomConf: "# Custom configuration\nclient_max_body_size 100M;",
	}

	config, err := GenerateConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v", err)
	}

	if !strings.Contains(config, "client_max_body_size 100M") {
		t.Error("Custom configuration not included")
	}
}

func TestGenerateSitePath(t *testing.T) {
	path := GenerateSitePath("myapp")
	expected := "/etc/nginx/sites-available/myapp"
	if path != expected {
		t.Errorf("GenerateSitePath() = %v, want %v", path, expected)
	}
}

func TestGenerateSymlinkPath(t *testing.T) {
	path := GenerateSymlinkPath("myapp")
	expected := "/etc/nginx/sites-enabled/myapp"
	if path != expected {
		t.Errorf("GenerateSymlinkPath() = %v, want %v", path, expected)
	}
}

func TestGenerateEnableCommand(t *testing.T) {
	cmd := GenerateEnableCommand("myapp")
	if !strings.Contains(cmd, "ln -sf") {
		t.Error("Enable command should use ln -sf")
	}
	if !strings.Contains(cmd, "sites-available/myapp") {
		t.Error("Enable command should reference sites-available")
	}
	if !strings.Contains(cmd, "sites-enabled/myapp") {
		t.Error("Enable command should reference sites-enabled")
	}
}

func TestGenerateReloadCommand(t *testing.T) {
	cmd := GenerateReloadCommand()
	if !strings.Contains(cmd, "nginx -t") {
		t.Error("Reload should test config first")
	}
	if !strings.Contains(cmd, "reload nginx") {
		t.Error("Reload should reload nginx")
	}
}

func TestGenerateTestCommand(t *testing.T) {
	cmd := GenerateTestCommand()
	expected := "sudo nginx -t"
	if cmd != expected {
		t.Errorf("GenerateTestCommand() = %v, want %v", cmd, expected)
	}
}

func TestGenerateInstallCommand(t *testing.T) {
	cmd := GenerateInstallCommand()
	if !strings.Contains(cmd, "apt-get install") {
		t.Error("Install command should use apt-get")
	}
	if !strings.Contains(cmd, "nginx") {
		t.Error("Install command should install nginx")
	}
}

