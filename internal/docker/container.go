package docker

import (
	"fmt"
	"strings"
)

// ContainerConfig represents Docker container configuration
type ContainerConfig struct {
	Name        string
	Image       string
	Port        int
	InternalPort int
	Env         map[string]string
	Labels      map[string]string
	Command     string
	Volumes     []string
	Options     map[string]string
}

// GenerateRunCommand generates a docker run command
func GenerateRunCommand(cfg ContainerConfig) string {
	var parts []string
	
	parts = append(parts, "sudo docker run -d")
	parts = append(parts, fmt.Sprintf("--name %s", cfg.Name))

	// Port mapping
	if cfg.Port > 0 {
		internalPort := cfg.InternalPort
		if internalPort == 0 {
			internalPort = cfg.Port
		}
		parts = append(parts, fmt.Sprintf("-p %d:%d", cfg.Port, internalPort))
	}

	// Environment variables
	for key, value := range cfg.Env {
		// Escape value for shell
		escapedValue := strings.ReplaceAll(value, `"`, `\"`)
		parts = append(parts, fmt.Sprintf(`-e %s="%s"`, key, escapedValue))
	}

	// Labels
	for key, value := range cfg.Labels {
		parts = append(parts, fmt.Sprintf(`--label %s=%s`, key, value))
	}

	// Volumes
	for _, vol := range cfg.Volumes {
		parts = append(parts, fmt.Sprintf("-v %s", vol))
	}

	// Custom options
	for key, value := range cfg.Options {
		if value == "" {
			parts = append(parts, fmt.Sprintf("--%s", key))
		} else {
			parts = append(parts, fmt.Sprintf("--%s=%s", key, value))
		}
	}

	// Image
	parts = append(parts, cfg.Image)

	// Command (optional)
	if cfg.Command != "" {
		parts = append(parts, cfg.Command)
	}

	return strings.Join(parts, " ")
}

// GenerateLoadCommand generates command to load Docker image from tar
func GenerateLoadCommand(tarPath string) string {
	return fmt.Sprintf("sudo docker load -i %s", tarPath)
}

// GenerateStopCommand generates command to stop container
func GenerateStopCommand(containerName string) string {
	return fmt.Sprintf("docker stop %s", containerName)
}

// GenerateRemoveCommand generates command to remove container
func GenerateRemoveCommand(containerName string) string {
	return fmt.Sprintf("docker rm %s", containerName)
}

// GenerateLogsCommand generates command to view logs
func GenerateLogsCommand(containerName string, tail int, follow bool) string {
	cmd := fmt.Sprintf("docker logs")
	
	if tail > 0 {
		cmd += fmt.Sprintf(" --tail %d", tail)
	}
	
	if follow {
		cmd += " -f"
	}
	
	cmd += " " + containerName
	
	return cmd
}

// GeneratePsCommand generates command to list containers
func GeneratePsCommand(serviceName string) string {
	return fmt.Sprintf(`docker ps --filter "label=podlift.service=%s" --format "{{.Names}}\t{{.Status}}\t{{.Label \"podlift.version\"}}"`, serviceName)
}

