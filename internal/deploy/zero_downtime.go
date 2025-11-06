package deploy

import (
	"fmt"
	"strings"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/docker"
	"github.com/ekinertac/podlift/internal/nginx"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// ZeroDowntimeDeployOptions contains options for zero-downtime deployment
type ZeroDowntimeDeployOptions struct {
	Config      *config.Config
	Version     string
	ImagePath   string
	SSHClient   ssh.SSHClient
	Server      config.Server
}

// ZeroDowntimeDeploy performs zero-downtime deployment with nginx
func ZeroDowntimeDeploy(opts ZeroDowntimeDeployOptions) error {
	cfg := opts.Config
	version := opts.Version
	client := opts.SSHClient

	fmt.Println(ui.Info("Starting zero-downtime deployment..."))

	// Step 1: Get existing containers
	existing, _, _ := client.CheckExistingService(cfg.Service)
	var oldContainers []string
	if existing {
		// Will stop these later
		psCmd := fmt.Sprintf(`sudo docker ps --filter "label=podlift.service=%s" --format "{{.Names}}"`, cfg.Service)
		output, _ := client.Execute(psCmd)
		if output != "" {
			for _, name := range strings.Split(strings.TrimSpace(output), "\n") {
				if name != "" {
					oldContainers = append(oldContainers, name)
				}
			}
		}
	}

	// Step 2: Start new containers on temp ports
	fmt.Println(ui.Info("Starting new containers..."))
	
	var newUpstreams []nginx.Upstream
	tempPortStart := 9000

	for serviceName, service := range cfg.Services {
		for replica := 1; replica <= service.Replicas; replica++ {
			containerName := fmt.Sprintf("%s-%s-%s-%d", cfg.Service, serviceName, version, replica)
			tempPort := tempPortStart + replica - 1

			containerCfg := docker.ContainerConfig{
				Name:         containerName,
				Image:        fmt.Sprintf("%s:%s", cfg.Image, version),
				Port:         tempPort,
				InternalPort: service.Port,
				Env:          service.Env,
				Labels: map[string]string{
					"podlift.service":       cfg.Service,
					"podlift.version":       version,
					"podlift.deployed_at":   time.Now().Format(time.RFC3339),
					"podlift.container_type": serviceName,
				},
				Command: service.Command,
				Volumes: service.Volumes,
			}

			runCmd := docker.GenerateRunCommand(containerCfg)
			if _, err := client.Execute(runCmd); err != nil {
				return fmt.Errorf("failed to start container %s: %w", containerName, err)
			}

			fmt.Println(ui.Success(fmt.Sprintf("  %s started on :%d", containerName, tempPort)))

			// Add to upstream list
			newUpstreams = append(newUpstreams, nginx.Upstream{
				Name: containerName,
				Host: "localhost",
				Port: tempPort,
			})
		}
	}

	// Step 3: Wait and health check new containers
	fmt.Println(ui.Info("Health checking new containers..."))
	time.Sleep(3 * time.Second)

	for _, service := range cfg.Services {
		if service.Healthcheck == nil || (service.Healthcheck.Enabled != nil && !*service.Healthcheck.Enabled) {
			continue
		}

		healthPath := service.Healthcheck.Path
		if healthPath == "" {
			healthPath = "/health"
		}

		// Check first new container
		firstPort := tempPortStart

		healthCfg := docker.HealthCheckConfig{
			URL:      fmt.Sprintf("http://%s:%d%s", opts.Server.Host, firstPort, healthPath),
			Expect:   service.Healthcheck.Expect,
			Timeout:  5 * time.Second,
			Interval: 2 * time.Second,
			Retries:  15,
		}

		if err := docker.CheckHealth(healthCfg); err != nil {
			fmt.Println(ui.Error("Health check failed on new containers"))
			fmt.Println(ui.Warning("Rolling back (stopping new containers)..."))
			
			// Rollback: stop new containers
			for _, upstream := range newUpstreams {
				client.Execute(fmt.Sprintf("sudo docker stop %s && sudo docker rm %s", upstream.Name, upstream.Name))
			}
			
			return fmt.Errorf("health check failed: %w", err)
		}

		fmt.Println(ui.Success("  New containers healthy"))
	}

	// Step 4: Update nginx upstream
	fmt.Println(ui.Info("Updating nginx configuration..."))
	
	nginxMgr := nginx.NewManager(client)

	// Ensure nginx is installed
	if installed, _ := nginxMgr.IsInstalled(); !installed {
		if err := nginxMgr.Install(); err != nil {
			return err
		}
	}

	// Update upstream to point to new containers
	sslCfg := nginx.SSLConfig{Enabled: false}
	if cfg.Proxy != nil && cfg.Proxy.SSL != "" && cfg.Proxy.SSL != "false" {
		// SSL configuration will be added in Phase 3
		sslCfg.Enabled = false // For now
	}

	domain := cfg.Domain
	if domain == "" {
		domain = opts.Server.Host
	}

	if err := nginxMgr.UpdateUpstream(cfg.Service, newUpstreams, domain, sslCfg); err != nil {
		return fmt.Errorf("failed to update nginx: %w", err)
	}

	fmt.Println(ui.Success("nginx updated to new containers"))

	// Step 5: Wait for connection draining (give nginx time to finish old requests)
	if len(oldContainers) > 0 {
		fmt.Println(ui.Info("Draining connections (5s)..."))
		time.Sleep(5 * time.Second)
	}

	// Step 6: Stop old containers
	if len(oldContainers) > 0 {
		fmt.Println(ui.Info("Stopping old containers..."))
		for _, containerName := range oldContainers {
			stopCmd := fmt.Sprintf("sudo docker stop %s && sudo docker rm %s", containerName, containerName)
			client.Execute(stopCmd)
			fmt.Println(ui.Info(fmt.Sprintf("  Stopped %s", containerName)))
		}
		fmt.Println(ui.Success("Old containers removed"))
	}

	return nil
}

