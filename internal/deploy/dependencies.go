package deploy

import (
	"fmt"
	"strings"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/docker"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// DeployDependencies deploys dependency containers (postgres, redis, etc.)
func DeployDependencies(cfg *config.Config, client ssh.SSHClient, version string) error {
	if len(cfg.Dependencies) == 0 {
		return nil // No dependencies
	}

	fmt.Println(ui.Info(fmt.Sprintf("Starting %d dependencies...", len(cfg.Dependencies))))
	fmt.Println()

	for name, dep := range cfg.Dependencies {
		// Check if dependency already running
		checkCmd := fmt.Sprintf(
			`sudo docker ps --filter "name=%s-%s" --format "{{.Names}}"`,
			cfg.Service, name,
		)
		
		output, _ := client.Execute(checkCmd)
		if strings.TrimSpace(output) != "" {
			fmt.Println(ui.Success(fmt.Sprintf("  %s: already running", name)))
			continue
		}

		// Start dependency
		fmt.Println(ui.Info(fmt.Sprintf("  Starting %s...", name)))

		containerName := fmt.Sprintf("%s-%s", cfg.Service, name)

		// Build docker run command
		containerCfg := docker.ContainerConfig{
			Name:    containerName,
			Image:   dep.Image,
			Port:    dep.Port,
			Env:     dep.Env,
			Command: dep.Command,
			Labels: map[string]string{
				"podlift.service":    cfg.Service,
				"podlift.dependency": name,
				"podlift.version":    version,
				"podlift.created_at": time.Now().Format(time.RFC3339),
			},
			Options: dep.Options,
		}

		// Handle volume
		if dep.Volume != "" {
			containerCfg.Volumes = []string{dep.Volume}
			
			// Create named volume if needed
			volumeName := strings.Split(dep.Volume, ":")[0]
			createVolumeCmd := fmt.Sprintf("sudo docker volume create %s 2>/dev/null || true", volumeName)
			client.Execute(createVolumeCmd)
		}

		// Add restart policy for dependencies
		if containerCfg.Options == nil {
			containerCfg.Options = make(map[string]string)
		}
		containerCfg.Options["restart"] = "unless-stopped"

		runCmd := docker.GenerateRunCommand(containerCfg)
		
		if _, err := client.Execute(runCmd); err != nil {
			return fmt.Errorf("failed to start dependency %s: %w", name, err)
		}

		fmt.Println(ui.Success(fmt.Sprintf("  %s: started", name)))

		// Wait for dependency to be healthy
		if err := waitForDependencyHealth(client, containerName, name, 30); err != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("  %s: %v", name, err)))
			fmt.Println(ui.Info(fmt.Sprintf("  %s: continuing anyway (check logs later)", name)))
		} else {
			fmt.Println(ui.Success(fmt.Sprintf("  %s: healthy", name)))
		}
	}

	fmt.Println()
	fmt.Println(ui.Success("All dependencies started"))
	fmt.Println()

	return nil
}

// StopDependencies stops dependency containers (used during removal)
func StopDependencies(cfg *config.Config, client ssh.SSHClient) error {
	if len(cfg.Dependencies) == 0 {
		return nil
	}

	fmt.Println(ui.Info("Stopping dependencies..."))

	for name := range cfg.Dependencies {
		containerName := fmt.Sprintf("%s-%s", cfg.Service, name)
		
		stopCmd := fmt.Sprintf("sudo docker stop %s && sudo docker rm %s", containerName, containerName)
		client.Execute(stopCmd) // Ignore errors
		
		fmt.Println(ui.Info(fmt.Sprintf("  %s: stopped", name)))
	}

	return nil
}

// CheckDependencies checks if dependencies are healthy
func CheckDependencies(cfg *config.Config, client ssh.SSHClient) (map[string]bool, error) {
	status := make(map[string]bool)

	for name := range cfg.Dependencies {
		containerName := fmt.Sprintf("%s-%s", cfg.Service, name)
		
		checkCmd := fmt.Sprintf(
			`sudo docker ps --filter "name=%s" --format "{{.Status}}"`,
			containerName,
		)
		
		output, err := client.Execute(checkCmd)
		if err != nil || !strings.Contains(output, "Up") {
			status[name] = false
		} else {
			status[name] = true
		}
	}

	return status, nil
}

// waitForDependencyHealth waits for a dependency to become healthy
func waitForDependencyHealth(client ssh.SSHClient, containerName, depName string, timeoutSec int) error {
	fmt.Println(ui.Info(fmt.Sprintf("  %s: waiting for health check...", depName)))
	
	startTime := time.Now()
	timeout := time.Duration(timeoutSec) * time.Second

	for {
		// Check if container is running
		checkCmd := fmt.Sprintf(
			`sudo docker inspect --format='{{.State.Status}}' %s 2>/dev/null`,
			containerName,
		)
		
		status, err := client.Execute(checkCmd)
		if err != nil {
			return fmt.Errorf("container not found")
		}

		status = strings.TrimSpace(status)

		// Check if container exited (failed to start)
		if status == "exited" || status == "dead" {
			return fmt.Errorf("container failed to start (status: %s)", status)
		}

		// If running, check health status
		if status == "running" {
			healthCmd := fmt.Sprintf(
				`sudo docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' %s`,
				containerName,
			)
			
			healthStatus, _ := client.Execute(healthCmd)
			healthStatus = strings.TrimSpace(healthStatus)

			// If no health check defined, assume healthy after running
			if healthStatus == "none" {
				time.Sleep(2 * time.Second) // Give it a moment to initialize
				return nil
			}

			// If health check passes
			if healthStatus == "healthy" {
				return nil
			}

			// If explicitly unhealthy (failed health check)
			if healthStatus == "unhealthy" {
				return fmt.Errorf("health check failed")
			}

			// Otherwise, keep waiting (status is "starting")
		}

		// Check timeout
		if time.Since(startTime) > timeout {
			return fmt.Errorf("health check timeout (%ds)", timeoutSec)
		}

		// Wait before retry
		time.Sleep(2 * time.Second)
	}
}

