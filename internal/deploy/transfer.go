package deploy

import (
	"fmt"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/docker"
	"github.com/ekinertac/podlift/internal/registry"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// transferImage transfers the Docker image to the server (via registry or SCP)
func transferImage(client ssh.SSHClient, host string, cfg *config.Config, version, tarPath string, opts DeployOptions) error {
	useRegistry := registry.IsConfigured(cfg)

	if useRegistry {
		// Use registry
		fmt.Println(ui.Info("Using registry for image transfer"))
		
		regClient := registry.NewClient(cfg.Registry)
		
		// Login on server
		if !opts.DryRun {
			if err := regClient.LoginRemote(client); err != nil {
				return err
			}

			// Pull image
			if err := regClient.Pull(client, cfg.Image, version); err != nil {
				return err
			}
		}

		fmt.Println(ui.Success("Image pulled from registry"))
		return nil
	}

	// Use SCP
	fmt.Println(ui.Info("Using SCP for image transfer"))
	remoteTarPath := fmt.Sprintf("/tmp/%s-%s.tar", cfg.Image, version)

	if !opts.DryRun {
		// Transfer with progress
		size, _ := docker.GetImageSize(tarPath)
		fmt.Println(ui.Info(fmt.Sprintf("  Uploading %.1fMB...", size)))
		
		err := client.CopyFileWithProgress(tarPath, remoteTarPath, func(sent, total int64) {
			percent := float64(sent) / float64(total) * 100
			if int(percent)%10 == 0 {
				fmt.Printf("\r  %.0f%% uploaded...", percent)
			}
		})
		if err != nil {
			return fmt.Errorf("failed to transfer image: %w", err)
		}
		fmt.Println()
	}

	fmt.Println(ui.Success("Image transferred"))

	// Load image on server
	fmt.Println(ui.Info("Loading image on server..."))
	
	if !opts.DryRun {
		loadCmd := docker.GenerateLoadCommand(remoteTarPath)
		if _, err := client.Execute(loadCmd); err != nil {
			return fmt.Errorf("failed to load image: %w", err)
		}

		// Remove tar file
		client.Execute(fmt.Sprintf("rm %s", remoteTarPath))
	}

	fmt.Println(ui.Success("Image loaded"))
	return nil
}

