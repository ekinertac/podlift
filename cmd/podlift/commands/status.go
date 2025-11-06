package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current deployment status",
	Long:  `Show current deployment status including version, services, and server health.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load("podlift.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.Title("Deployment Status"))
	fmt.Println()

	// Get all servers
	allServers := cfg.GetAllServers()

	// Connect to primary server to get current version
	primaryServer, _, err := cfg.GetPrimaryServer()
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(ssh.Config{
		Host:    primaryServer.Host,
		Port:    primaryServer.Port,
		User:    primaryServer.User,
		KeyPath: primaryServer.SSHKey,
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Get current deployment version
	fmt.Println(ui.Info("Current deployment:"))
	
	// List containers
	listCmd := fmt.Sprintf("sudo docker ps --filter label=podlift.service=%s --format '{{.Names}}\t{{.Image}}\t{{.Status}}'", cfg.Service)
	output, err := client.Execute(listCmd)
	if err != nil {
		fmt.Println(ui.Error("  No deployment found"))
	} else if output == "" {
		fmt.Println(ui.Error("  No containers running"))
	} else {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) > 0 {
			// Extract version from first container image
			parts := strings.Split(lines[0], "\t")
			if len(parts) >= 2 {
				imageParts := strings.Split(parts[1], ":")
				if len(imageParts) >= 2 {
					version := imageParts[len(imageParts)-1]
					fmt.Println(ui.Success(fmt.Sprintf("  Version: %s", version)))
				}
			}
			
			healthyCount := 0
			totalCount := len(lines)
			for _, line := range lines {
				if strings.Contains(line, "healthy") || strings.Contains(line, "Up") {
					healthyCount++
				}
			}
			
			if healthyCount == totalCount {
				fmt.Println(ui.Success(fmt.Sprintf("  Services: %d/%d healthy", healthyCount, totalCount)))
			} else {
				fmt.Println(ui.Warning(fmt.Sprintf("  Services: %d/%d healthy", healthyCount, totalCount)))
			}
		}
	}

	if cfg.Domain != "" {
		protocol := "http"
		if cfg.Proxy != nil && cfg.Proxy.SSL != "" {
			protocol = "https"
		}
		fmt.Println(ui.Info(fmt.Sprintf("  URL: %s://%s", protocol, cfg.Domain)))
	}
	fmt.Println()

	// Show servers status
	fmt.Println(ui.Info("Servers:"))
	for _, server := range allServers {
		srvClient, err := ssh.NewClient(ssh.Config{
			Host:    server.Host,
			Port:    server.Port,
			User:    server.User,
			KeyPath: server.SSHKey,
			Timeout: 10 * time.Second,
		})
		if err != nil {
			fmt.Println(ui.Error(fmt.Sprintf("  %s - connection failed", server.Host)))
			continue
		}

		if err := srvClient.Connect(); err != nil {
			fmt.Println(ui.Error(fmt.Sprintf("  %s - connection failed", server.Host)))
			srvClient.Close()
			continue
		}

		// Count containers on this server
		countCmd := fmt.Sprintf("sudo docker ps --filter label=podlift.service=%s --format '{{.Names}}' | wc -l", cfg.Service)
		countOutput, err := srvClient.Execute(countCmd)
		containerCount := "0"
		if err == nil {
			containerCount = strings.TrimSpace(countOutput)
		}

		// Check for dependencies
		depCmd := "sudo docker ps --filter label=podlift.dependency=true --format '{{.Names}}' | sed 's/.*-//'"
		depOutput, err := srvClient.Execute(depCmd)
		deps := ""
		if err == nil && depOutput != "" {
			depList := strings.Split(strings.TrimSpace(depOutput), "\n")
			deps = strings.Join(depList, ", ")
		}

		status := fmt.Sprintf("  %s - healthy (%s containers", server.Host, containerCount)
		if deps != "" {
			status += fmt.Sprintf(", %s", deps)
		}
		status += ")"
		
		fmt.Println(ui.Success(status))
		srvClient.Close()
	}
	fmt.Println()

	// Show available commands
	fmt.Println(ui.Info("Available commands:"))
	for serviceName := range cfg.Services {
		fmt.Println(fmt.Sprintf("  podlift logs %s", serviceName))
		fmt.Println(fmt.Sprintf("  podlift exec %s bash", serviceName))
	}
	fmt.Println("  podlift rollback")
	fmt.Println()

	return nil
}

