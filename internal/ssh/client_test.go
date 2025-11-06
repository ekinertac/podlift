package ssh

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Note: This test requires a valid SSH key to exist
	// It only tests client creation, not connection

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:    "localhost",
				Port:    22,
				User:    "test",
				KeyPath: "~/.ssh/id_rsa",
				Timeout: 30 * time.Second,
			},
			wantErr: false, // May fail if key doesn't exist, but that's expected
		},
		{
			name: "defaults applied",
			config: Config{
				Host:    "localhost",
				User:    "test",
				KeyPath: "~/.ssh/id_rsa",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			
			// If key doesn't exist, that's okay for this test
			if err != nil && !tt.wantErr {
				// Check if it's a "file not found" error (expected if no SSH key)
				if _, ok := err.(*os.PathError); ok {
					t.Skipf("SSH key not found (expected): %v", err)
				}
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if client == nil {
					t.Error("NewClient() returned nil client")
				}
				if client.host != tt.config.Host {
					t.Errorf("Client host = %v, want %v", client.host, tt.config.Host)
				}
				if tt.config.Port != 0 && client.port != tt.config.Port {
					t.Errorf("Client port = %v, want %v", client.port, tt.config.Port)
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		check func(string) bool
	}{
		{
			name: "tilde expansion",
			path: "~/.ssh/id_rsa",
			check: func(expanded string) bool {
				return len(expanded) > 0 && expanded[0] != '~'
			},
		},
		{
			name: "absolute path unchanged",
			path: "/etc/ssh/id_rsa",
			check: func(expanded string) bool {
				return expanded == "/etc/ssh/id_rsa"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.path)
			if !tt.check(result) {
				t.Errorf("expandPath(%v) = %v, check failed", tt.path, result)
			}
		})
	}
}

// Integration tests below require SSH server
// Skip if PODLIFT_SSH_TEST_HOST is not set

func getTestHost() string {
	return os.Getenv("PODLIFT_SSH_TEST_HOST")
}

func TestConnect_Integration(t *testing.T) {
	host := getTestHost()
	if host == "" {
		t.Skip("Skipping integration test: set PODLIFT_SSH_TEST_HOST to enable")
	}

	client, err := NewClient(Config{
		Host:    host,
		Port:    22,
		User:    "root",
		KeyPath: "~/.ssh/id_rsa",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	err = client.Connect()
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if !client.connected {
		t.Error("client.connected = false after Connect()")
	}
}

func TestExecute_Integration(t *testing.T) {
	host := getTestHost()
	if host == "" {
		t.Skip("Skipping integration test: set PODLIFT_SSH_TEST_HOST to enable")
	}

	client, err := NewClient(Config{
		Host:    host,
		User:    "root",
		KeyPath: "~/.ssh/id_rsa",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	output, err := client.Execute("echo 'hello world'")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(output, "hello world") {
		t.Errorf("Execute() output = %v, want to contain 'hello world'", output)
	}
}

func TestCheckDocker_Integration(t *testing.T) {
	host := getTestHost()
	if host == "" {
		t.Skip("Skipping integration test: set PODLIFT_SSH_TEST_HOST to enable")
	}

	client, err := NewClient(Config{
		Host:    host,
		User:    "root",
		KeyPath: "~/.ssh/id_rsa",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	version, err := client.CheckDocker()
	if err != nil {
		t.Fatalf("CheckDocker() error = %v", err)
	}

	if !strings.Contains(version, "Docker version") {
		t.Errorf("CheckDocker() = %v, want to contain 'Docker version'", version)
	}
}

