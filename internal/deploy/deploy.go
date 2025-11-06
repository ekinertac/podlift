package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/docker"
	"github.com/ekinertac/podlift/internal/git"
	"github.com/ekinertac/podlift/internal/registry"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// DeployOptions contains deployment configuration
type DeployOptions struct {
	Config       *config.Config
	SkipBuild    bool
	SkipHealth   bool
	Parallel     bool
	DryRun       bool
	ZeroDowntime bool
}

// Deploy executes a deployment
func Deploy(opts DeployOptions) error {
	cfg := opts.Config

	// Get version from git
	version, err := git.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	fmt.Println(ui.Title(fmt.Sprintf("Deploying %s:%s", cfg.Service, version)))
	fmt.Println()

	// Determine transfer method
	useRegistry := registry.IsConfigured(cfg)
	transferMethod := "SCP"
	if useRegistry {
		transferMethod = "Registry"
	}

	steps := ui.NewStepList([]string{
		"Build image",
		fmt.Sprintf("Transfer (%s)", transferMethod),
		"Load/Pull image",
		"Start containers",
		"Health check",
	})

	// Step 1: Build image
	steps.Start(0, fmt.Sprintf("Building %s:%s...", cfg.Image, version))
	fmt.Println(steps.RenderCurrent())

	if !opts.SkipBuild && !opts.DryRun {
		cwd, _ := os.Getwd()
		if err := docker.BuildImage(cfg.Image, version, cwd); err != nil {
			steps.Fail(0, err.Error())
			return err
		}
	}

	steps.Complete(0, "Built successfully")
	fmt.Println(ui.Success("Build complete"))
	fmt.Println()

	// Step 2: Push to registry or save to tar
	var tarPath string
	
	if useRegistry {
		// Push to registry
		steps.Start(1, "Pushing to registry...")
		fmt.Println(steps.RenderCurrent())

		if !opts.DryRun {
			regClient := registry.NewClient(cfg.Registry)
			
			// Login
			if err := regClient.Login(); err != nil {
				steps.Fail(1, err.Error())
				return err
			}

			// Push
			if err := regClient.Push(cfg.Image, version); err != nil {
				steps.Fail(1, err.Error())
				return err
			}
		}

		steps.Complete(1, "Pushed to registry")
		fmt.Println(ui.Success("Image pushed"))
		fmt.Println()
	} else {
		// Save to tar for SCP
		tempDir := filepath.Join(os.TempDir(), "podlift")
		os.MkdirAll(tempDir, 0755)
		tarPath = filepath.Join(tempDir, fmt.Sprintf("%s-%s.tar", cfg.Image, version))

		steps.Start(1, "Saving image to tar...")
		fmt.Println(steps.RenderCurrent())

		if !opts.DryRun {
			if err := docker.SaveImage(cfg.Image, version, tarPath); err != nil {
				steps.Fail(1, err.Error())
				return err
			}
			defer os.Remove(tarPath) // Cleanup
		}

		// Get image size
		size, _ := docker.GetImageSize(tarPath)
		steps.Complete(1, fmt.Sprintf("Saved (%.1f MB)", size))
		fmt.Println(ui.Success(fmt.Sprintf("Image saved: %.1fMB", size)))
		fmt.Println()
	}

	// Deploy to each server
	allServers := cfg.GetAllServers()
	
	for i, serverWithRole := range allServers {
		fmt.Printf("Server %d/%d: %s\n", i+1, len(allServers), serverWithRole.Host)
		fmt.Println()

		// Create SSH client for this server
		sshClient, err := ssh.NewClient(ssh.Config{
			Host:    serverWithRole.Host,
			Port:    serverWithRole.Port,
			User:    serverWithRole.User,
			KeyPath: serverWithRole.SSHKey,
			Timeout: 30 * time.Second,
		})
		if err != nil {
			return fmt.Errorf("failed to create SSH client: %w", err)
		}
		defer sshClient.Close()

		if err := sshClient.Connect(); err != nil {
			return fmt.Errorf("SSH connection failed: %w", err)
		}

		// Transfer image
		if err := transferImage(sshClient, serverWithRole.Host, cfg, version, tarPath, opts); err != nil {
			return err
		}

		// Deploy dependencies first (postgres, redis, etc.)
		if !opts.DryRun {
			if err := DeployDependencies(cfg, sshClient, version); err != nil {
				return fmt.Errorf("dependency deployment failed: %w", err)
			}
		}

		// Choose deployment strategy
		if opts.ZeroDowntime {
			// Zero-downtime deployment with nginx
			zdOpts := ZeroDowntimeDeployOptions{
				Config:    cfg,
				Version:   version,
				ImagePath: tarPath,
				SSHClient: sshClient,
				Server:    serverWithRole.Server,
			}
			if err := ZeroDowntimeDeploy(zdOpts); err != nil {
				return fmt.Errorf("deployment failed on %s: %w", serverWithRole.Host, err)
			}
		} else {
			// Basic deployment (current method)
			if err := deployToServer(serverWithRole.Server, cfg, version, tarPath, opts); err != nil {
				return fmt.Errorf("deployment failed on %s: %w", serverWithRole.Host, err)
			}
		}

		fmt.Println(ui.Success(fmt.Sprintf("Deployed to %s", serverWithRole.Host)))
		fmt.Println()
	}

	fmt.Println(ui.Title("Deployment successful!"))
	fmt.Println()
	fmt.Println(ui.Info(fmt.Sprintf("Deployed: %s", version)))
	
	commitMsg, _ := git.GetCommitMessage()
	if commitMsg != "" {
		fmt.Println(ui.Info(fmt.Sprintf("Message: %s", commitMsg)))
	}

	return nil
}

// deployToServer deploys to a single server
func deployToServer(server config.Server, cfg *config.Config, version, tarPath string, opts DeployOptions) error {
	// Create SSH client
	sshClient, err := ssh.NewClient(ssh.Config{
		Host:    server.Host,
		Port:    server.Port,
		User:    server.User,
		KeyPath: server.SSHKey,
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return err
	}
	defer sshClient.Close()

	// Connect
	if err := sshClient.Connect(); err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	// Step 3: Transfer tar to server
	fmt.Println(ui.Info("Transferring image..."))
	remoteTarPath := fmt.Sprintf("/tmp/%s-%s.tar", cfg.Image, version)

	if !opts.DryRun {
		// Transfer with progress
		size, _ := docker.GetImageSize(tarPath)
		fmt.Println(ui.Info(fmt.Sprintf("  Uploading %.1fMB...", size)))
		
		err := sshClient.CopyFileWithProgress(tarPath, remoteTarPath, func(sent, total int64) {
			// Progress callback - could show progress bar here
			percent := float64(sent) / float64(total) * 100
			if int(percent)%10 == 0 { // Show every 10%
				fmt.Printf("\r  %.0f%% uploaded...", percent)
			}
		})
		if err != nil {
			return fmt.Errorf("failed to transfer image: %w", err)
		}
		fmt.Println() // New line after progress
	}

	fmt.Println(ui.Success("Image transferred"))

	// Step 4: Load image on server
	fmt.Println(ui.Info("Loading image on server..."))
	
	if !opts.DryRun {
		loadCmd := docker.GenerateLoadCommand(remoteTarPath)
		if _, err := sshClient.Execute(loadCmd); err != nil {
			return fmt.Errorf("failed to load image: %w", err)
		}

		// Remove tar file
		sshClient.Execute(fmt.Sprintf("rm %s", remoteTarPath))
	}

	fmt.Println(ui.Success("Image loaded"))

	// Step 5: Start containers
	fmt.Println(ui.Info("Starting containers..."))

	for serviceName, service := range cfg.Services {
		for replica := 1; replica <= service.Replicas; replica++ {
			containerName := fmt.Sprintf("%s-%s-%s-%d", cfg.Service, serviceName, version, replica)
			
			// Generate docker run command
			containerCfg := docker.ContainerConfig{
				Name:         containerName,
				Image:        fmt.Sprintf("%s:%s", cfg.Image, version),
				Port:         8000 + replica - 1, // Temp ports
				InternalPort: service.Port,
				Env:          service.Env,
				Labels: map[string]string{
					"podlift.service":  cfg.Service,
					"podlift.version":  version,
					"podlift.deployed_at": time.Now().Format(time.RFC3339),
					"podlift.container_type": serviceName,
				},
				Command: service.Command,
				Volumes: service.Volumes,
			}

			runCmd := docker.GenerateRunCommand(containerCfg)

			if opts.DryRun {
				fmt.Println(ui.Code("  " + runCmd))
			} else {
				if _, err := sshClient.Execute(runCmd); err != nil {
					return fmt.Errorf("failed to start container %s: %w", containerName, err)
				}
				fmt.Println(ui.Success(fmt.Sprintf("  %s started", containerName)))
			}
		}
	}

	// Step 6: Health check
	if !opts.SkipHealth && !opts.DryRun {
		fmt.Println(ui.Info("Waiting for health check..."))
		
		// Check each service with health checks
		for serviceName, service := range cfg.Services {
			if service.Healthcheck == nil || service.Healthcheck.Enabled != nil && !*service.Healthcheck.Enabled {
				fmt.Println(ui.Info(fmt.Sprintf("  %s: health check disabled", serviceName)))
				continue
			}

			// Wait a moment for container to start
			time.Sleep(3 * time.Second)

			// Perform health check
			healthPath := service.Healthcheck.Path
			if healthPath == "" {
				healthPath = "/health"
			}

			expectedCodes := service.Healthcheck.Expect
			if len(expectedCodes) == 0 {
				expectedCodes = []int{200}
			}

			// Check first replica
			port := 8000 // Temp port for first replica
			
			healthCfg := docker.HealthCheckConfig{
				URL:      fmt.Sprintf("http://%s:%d%s", server.Host, port, healthPath),
				Expect:   expectedCodes,
				Timeout:  5 * time.Second,
				Interval: 2 * time.Second,
				Retries:  10,
			}

			if err := docker.CheckHealth(healthCfg); err != nil {
				return fmt.Errorf("health check failed for %s: %w", serviceName, err)
			}

			fmt.Println(ui.Success(fmt.Sprintf("  %s: healthy", serviceName)))
		}
	}

	return nil
}

