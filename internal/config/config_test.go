package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "minimal config",
			yaml: `
service: myapp
image: myapp
servers:
  - host: 192.168.1.10
`,
			wantErr: false,
		},
		{
			name: "missing service",
			yaml: `
image: myapp
servers:
  - host: 192.168.1.10
`,
			wantErr: true,
		},
		{
			name: "missing servers",
			yaml: `
service: myapp
image: myapp
`,
			wantErr: true,
		},
		{
			name: "full config",
			yaml: `
service: myapp
domain: myapp.com
image: myapp
servers:
  web:
    - host: 192.168.1.10
      user: root
      labels: [primary]
  worker:
    - host: 192.168.1.12
registry:
  server: ghcr.io
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASSWORD}
dependencies:
  postgres:
    image: postgres:16
    port: 5432
services:
  web:
    port: 8000
    replicas: 2
    healthcheck:
      path: /health
      expect: [200]
proxy:
  enabled: true
  ssl: letsencrypt
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write to temp file
			tmpfile, err := os.CreateTemp("", "podlift-test-*.yml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.WriteString(tt.yaml); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

			// Load config
			config, err := Load(tmpfile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config == nil {
					t.Error("Load() returned nil config")
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Service: "myapp",
				Image:   "myapp",
				Servers: ServersConfig{servers: map[string][]Server{
					"web": {{Host: "192.168.1.10"}},
				}},
			},
			wantErr: false,
		},
		{
			name: "missing service",
			config: Config{
				Image: "myapp",
				Servers: ServersConfig{servers: map[string][]Server{
					"web": {{Host: "192.168.1.10"}},
				}},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Service: "myapp",
				Image:   "myapp",
				Servers: ServersConfig{servers: map[string][]Server{
					"web": {{Host: "192.168.1.10"}},
				}},
				Services: map[string]Service{
					"web": {Port: 70000}, // Invalid port
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.applyDefaults()
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPrimaryServer(t *testing.T) {
	config := Config{
		Service: "myapp",
		Image:   "myapp",
		Servers: ServersConfig{servers: map[string][]Server{
			"web": {
				{Host: "192.168.1.10", Labels: []string{"primary"}},
				{Host: "192.168.1.11"},
			},
		}},
	}

	server, role, err := config.GetPrimaryServer()
	if err != nil {
		t.Fatalf("GetPrimaryServer() error = %v", err)
	}

	if server.Host != "192.168.1.10" {
		t.Errorf("GetPrimaryServer() host = %v, want 192.168.1.10", server.Host)
	}

	if role != "web" {
		t.Errorf("GetPrimaryServer() role = %v, want web", role)
	}
}

