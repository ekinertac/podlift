package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client connection
type Client struct {
	config     *ssh.ClientConfig
	host       string
	port       int
	client     *ssh.Client
	connected  bool
}

// Config represents SSH connection configuration
type Config struct {
	Host       string
	Port       int
	User       string
	KeyPath    string
	Timeout    time.Duration
}

// NewClient creates a new SSH client
func NewClient(cfg Config) (*Client, error) {
	// Set defaults
	if cfg.Port == 0 {
		cfg.Port = 22
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Read SSH private key
	keyPath := expandPath(cfg.KeyPath)
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH key %s: %w", keyPath, err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH key: %w", err)
	}

	// Setup known hosts
	// For now, accept any host key (InsecureIgnoreHostKey)
	// TODO: Make this configurable (--strict-host-key-checking flag)
	hostKeyCallback := ssh.InsecureIgnoreHostKey()

	// Create SSH config
	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         cfg.Timeout,
	}

	return &Client{
		config: sshConfig,
		host:   cfg.Host,
		port:   cfg.Port,
	}, nil
}

// Connect establishes the SSH connection
func (c *Client) Connect() error {
	if c.connected {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	client, err := ssh.Dial("tcp", addr, c.config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.client = client
	c.connected = true
	return nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Execute runs a command on the remote server
func (c *Client) Execute(command string) (string, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return "", err
		}
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("command failed: %w\nstdout: %s\nstderr: %s", 
			err, stdout.String(), stderr.String())
	}

	return stdout.String(), nil
}

// ExecuteWithOutput runs a command and streams output
func (c *Client) ExecuteWithOutput(command string, stdout, stderr io.Writer) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// TestConnection tests if SSH connection works
func (c *Client) TestConnection() error {
	if err := c.Connect(); err != nil {
		return err
	}

	// Run simple command to verify
	_, err := c.Execute("echo 'test'")
	return err
}

// CheckDocker verifies Docker is installed on the remote server
func (c *Client) CheckDocker() (string, error) {
	output, err := c.Execute("docker --version")
	if err != nil {
		return "", fmt.Errorf("Docker not installed or not in PATH: %w", err)
	}
	return output, nil
}

// CheckPort checks if a port is available on the remote server
func (c *Client) CheckPort(port int) (bool, error) {
	// Use netstat or ss to check if port is in use
	cmd := fmt.Sprintf("netstat -tuln 2>/dev/null | grep ':%d ' || ss -tuln 2>/dev/null | grep ':%d '", port, port)
	output, err := c.Execute(cmd)
	if err != nil {
		// Command failed, likely means port is free
		return true, nil
	}

	// If we got output, port is in use
	if len(output) > 0 {
		return false, nil
	}

	return true, nil
}

// GetDiskSpace returns available disk space in GB
func (c *Client) GetDiskSpace() (float64, error) {
	output, err := c.Execute("df -BG / | tail -1 | awk '{print $4}'")
	if err != nil {
		return 0, err
	}

	var space float64
	_, err = fmt.Sscanf(output, "%fG", &space)
	if err != nil {
		return 0, fmt.Errorf("failed to parse disk space: %w", err)
	}

	return space, nil
}

// expandPath expands ~ in paths
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}

