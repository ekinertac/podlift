package registry

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// Client handles Docker registry operations
type Client struct {
	config *config.RegistryConfig
}

// NewClient creates a new registry client
func NewClient(cfg *config.RegistryConfig) *Client {
	return &Client{config: cfg}
}

// Login logs into the Docker registry locally
func (c *Client) Login() error {
	if c.config == nil {
		return fmt.Errorf("registry not configured")
	}

	server := c.config.Server
	if server == "" {
		server = "docker.io"
	}

	fmt.Println(ui.Info(fmt.Sprintf("Logging into %s...", server)))

	// Docker login
	cmd := exec.Command("docker", "login", server, 
		"-u", c.config.Username,
		"--password-stdin")
	
	// Pass password via stdin (more secure than command line)
	cmd.Stdin = strings.NewReader(c.config.Password)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("registry login failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Println(ui.Success("Logged into registry"))
	return nil
}

// LoginRemote logs into registry on remote server
func (c *Client) LoginRemote(client ssh.SSHClient) error {
	if c.config == nil {
		return fmt.Errorf("registry not configured")
	}

	server := c.config.Server
	if server == "" {
		server = "docker.io"
	}

	fmt.Println(ui.Info(fmt.Sprintf("Logging into %s on server...", server)))

	// Use echo to pass password via stdin
	loginCmd := fmt.Sprintf(
		"echo '%s' | sudo docker login %s -u %s --password-stdin",
		c.config.Password, server, c.config.Username,
	)

	if _, err := client.Execute(loginCmd); err != nil {
		return fmt.Errorf("registry login failed on server: %w", err)
	}

	fmt.Println(ui.Success("Server logged into registry"))
	return nil
}

// Push pushes an image to the registry
func (c *Client) Push(imageName, tag string) error {
	if c.config == nil {
		return fmt.Errorf("registry not configured")
	}

	// Tag image for registry
	registryImage := c.GetImagePath(imageName, tag)
	
	fmt.Println(ui.Info(fmt.Sprintf("Tagging image as %s...", registryImage)))
	
	tagCmd := exec.Command("docker", "tag", fmt.Sprintf("%s:%s", imageName, tag), registryImage)
	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	// Push to registry
	fmt.Println(ui.Info(fmt.Sprintf("Pushing to %s...", c.config.Server)))
	
	pushCmd := exec.Command("docker", "push", registryImage)
	pushCmd.Stdout = &progressWriter{}
	pushCmd.Stderr = &progressWriter{}
	
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	fmt.Println(ui.Success("Image pushed to registry"))
	return nil
}

// Pull pulls an image from registry on remote server
func (c *Client) Pull(client ssh.SSHClient, imageName, tag string) error {
	if c.config == nil {
		return fmt.Errorf("registry not configured")
	}

	registryImage := c.GetImagePath(imageName, tag)
	
	fmt.Println(ui.Info(fmt.Sprintf("Pulling %s...", registryImage)))

	pullCmd := fmt.Sprintf("sudo docker pull %s", registryImage)
	output, err := client.Execute(pullCmd)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w\nOutput: %s", err, output)
	}

	fmt.Println(ui.Success("Image pulled from registry"))
	return nil
}

// GetImagePath returns the full registry path for an image
func (c *Client) GetImagePath(imageName, tag string) string {
	if c.config == nil {
		return fmt.Sprintf("%s:%s", imageName, tag)
	}

	server := c.config.Server
	if server == "" || server == "docker.io" {
		// Docker Hub format: username/image:tag
		return fmt.Sprintf("%s/%s:%s", c.config.Username, imageName, tag)
	}

	// Other registries: server/username/image:tag
	return fmt.Sprintf("%s/%s/%s:%s", server, c.config.Username, imageName, tag)
}

// IsConfigured checks if registry is configured
func IsConfigured(cfg *config.Config) bool {
	return cfg.Registry != nil && 
		cfg.Registry.Username != "" && 
		cfg.Registry.Password != ""
}

// progressWriter shows docker push/pull progress
type progressWriter struct{}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	output := strings.TrimSpace(string(p))
	if output != "" && !strings.Contains(output, "Waiting") {
		// Only show meaningful progress
		if strings.Contains(output, "Pushed") || strings.Contains(output, "Pulling") {
			fmt.Println(ui.Info("  " + output))
		}
	}
	return len(p), nil
}

