package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/nginx"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

var (
	rollbackTo             string
	rollbackSkipHealthcheck bool
)

func init() {
	rollbackCommand.Flags().StringVar(&rollbackTo, "to", "", "Rollback to specific git commit or tag")
	rollbackCommand.Flags().BoolVar(&rollbackSkipHealthcheck, "skip-healthcheck", false, "Don't wait for health checks")
	rootCmd.AddCommand(rollbackCommand)
}

var rollbackCommand = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to previous deployment",
	Long:  "Reverts to the previous version by restarting old containers",
	RunE:  runRollback,
}

func runRollback(cmd *cobra.Command, args []string) error {
	// Load configuration
	configPath, err := config.Find()
	if err != nil {
		return fmt.Errorf("podlift.yml not found")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	fmt.Println(ui.Title(fmt.Sprintf("Rolling back %s", cfg.Service)))
	fmt.Println()

	// Get all servers
	allServers := cfg.GetAllServers()

	// For each server, find previous containers and restart them
	for _, server := range allServers {
		fmt.Println(ui.Info(fmt.Sprintf("Rolling back on %s", server.Host)))

		client, err := ssh.NewClient(ssh.Config{
			Host:    server.Host,
			Port:    server.Port,
			User:    server.User,
			KeyPath: server.SSHKey,
			Timeout: 30 * time.Second,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", server.Host, err)
		}
		defer client.Close()

		if err := client.Connect(); err != nil {
			return err
		}

		// Find stopped containers with podlift.service label
		listCmd := fmt.Sprintf("sudo docker ps -a --filter label=podlift.service=%s --filter status=exited --format '{{.Names}}\t{{.Image}}' | head -10", cfg.Service)
		output, err := client.Execute(listCmd)
		if err != nil || output == "" {
			fmt.Println(ui.Warning("  No previous deployment found"))
			continue
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) == 0 {
			fmt.Println(ui.Warning("  No previous containers found"))
			continue
		}

		// If --to specified, filter by version
		var targetContainers []string
		if rollbackTo != "" {
			for _, line := range lines {
				if strings.Contains(line, rollbackTo) {
					parts := strings.Fields(line)
					if len(parts) > 0 {
						targetContainers = append(targetContainers, parts[0])
					}
				}
			}
		} else {
			// Use first (most recent) stopped containers
			for _, line := range lines {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					targetContainers = append(targetContainers, parts[0])
				}
			}
		}

		if len(targetContainers) == 0 {
			fmt.Println(ui.Warning("  No matching containers found"))
			continue
		}

		// Stop current running containers
		stopCmd := fmt.Sprintf("sudo docker ps --filter label=podlift.service=%s --format '{{.Names}}' | xargs -r sudo docker stop", cfg.Service)
		_, err = client.Execute(stopCmd)
		if err != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("  Failed to stop current containers: %v", err)))
		}

		// Start previous containers
		for _, container := range targetContainers {
			fmt.Println(ui.Info(fmt.Sprintf("  Starting %s", container)))
			startCmd := fmt.Sprintf("sudo docker start %s", container)
			_, err := client.Execute(startCmd)
			if err != nil {
				return fmt.Errorf("failed to start %s: %w", container, err)
			}
			fmt.Println(ui.Success(fmt.Sprintf("  ✓ %s started", container)))
		}

		// Update nginx if configured
		if cfg.Proxy != nil && cfg.Proxy.Enabled {
			nginxMgr := nginx.NewManager(client)
			// TODO: Update nginx upstream to point to rolled back containers
			_ = nginxMgr
		}
	}

	fmt.Println()
	fmt.Println(ui.Success("Rollback complete!"))
	fmt.Println()

	return nil
}

