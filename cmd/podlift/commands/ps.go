package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/docker"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

var psAll bool

func init() {
	rootCmd.AddCommand(psCommand)
	psCommand.Flags().BoolVarP(&psAll, "all", "a", false, "Show all containers (including stopped)")
}

var psCommand = &cobra.Command{
	Use:   "ps",
	Short: "Show running services",
	Long:  "Display status of running containers",
	RunE:  runPs,
}

func runPs(cmd *cobra.Command, args []string) error {
	// Load configuration
	configPath, err := config.Find()
	if err != nil {
		return fmt.Errorf("podlift.yml not found")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	fmt.Println(ui.Title(fmt.Sprintf("Services for: %s", cfg.Service)))
	fmt.Println()

	// Show dependencies first if any
	if len(cfg.Dependencies) > 0 {
		fmt.Println(ui.Info("Dependencies:"))
		for name, dep := range cfg.Dependencies {
			fmt.Println(ui.Info(fmt.Sprintf("  %s (%s)", name, dep.Image)))
		}
		fmt.Println()
	}

	// Get all servers
	allServers := cfg.GetAllServers()

	var allRows []table.Row

	for _, serverWithRole := range allServers {
		server := serverWithRole.Server

		// Create SSH client
		sshClient, err := ssh.NewClient(ssh.Config{
			Host:    server.Host,
			Port:    server.Port,
			User:    server.User,
			KeyPath: server.SSHKey,
			Timeout: 10 * time.Second,
		})
		if err != nil {
			return err
		}
		defer sshClient.Close()

		// List containers
		psCmd := docker.GeneratePsCommand(cfg.Service)
		if psAll {
			psCmd = fmt.Sprintf(`docker ps -a --filter "label=podlift.service=%s" --format "{{.Names}}\t{{.Status}}\t{{.Label \"podlift.version\"}}"`, cfg.Service)
		}

		output, err := sshClient.Execute(psCmd)
		if err != nil {
			fmt.Printf("Server %s: %v\n", server.Host, err)
			continue
		}

		output = strings.TrimSpace(output)
		if output == "" {
			allRows = append(allRows, table.Row{
				server.Host,
				"none",
				"-",
				"-",
				"-",
			})
			continue
		}

		// Parse output
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			parts := strings.Split(line, "\t")
			if len(parts) < 2 {
				continue
			}

			name := parts[0]
			status := parts[1]
			version := "-"
			if len(parts) >= 3 {
				version = parts[2]
			}

			// Determine health status
			healthStatus := "unknown"
			if strings.Contains(status, "Up") {
				if strings.Contains(status, "healthy") {
					healthStatus = "healthy"
				} else {
					healthStatus = "running"
				}
			} else {
				healthStatus = "stopped"
			}

			// Extract uptime from status
			uptime := extractUptime(status)

			allRows = append(allRows, table.Row{
				server.Host,
				extractServiceName(name),
				healthStatus,
				version,
				uptime,
			})
		}
	}

	// Display table
	if len(allRows) == 0 {
		fmt.Println(ui.Info("No containers found"))
		fmt.Println()
		fmt.Println("Deploy your application: podlift deploy")
		return nil
	}

	columns := []table.Column{
		{Title: "Server", Width: 20},
		{Title: "Container", Width: 25},
		{Title: "Status", Width: 12},
		{Title: "Version", Width: 10},
		{Title: "Uptime", Width: 15},
	}

	tbl := ui.NewTable(columns, allRows)
	fmt.Println(tbl.Render())
	fmt.Println()

	return nil
}

// extractServiceName extracts service name from container name
// e.g., "myapp-web-abc123-1" → "web"
func extractServiceName(containerName string) string {
	parts := strings.Split(containerName, "-")
	if len(parts) >= 2 {
		return parts[1] // Return "web" from "myapp-web-abc123-1"
	}
	return containerName
}

// extractUptime extracts uptime from Docker status string
// e.g., "Up 5 minutes" → "5m"
func extractUptime(status string) string {
	if !strings.Contains(status, "Up ") {
		return "-"
	}

	// Simple extraction
	if strings.Contains(status, "second") {
		return "<1m"
	}
	if strings.Contains(status, "minute") {
		parts := strings.Split(status, " ")
		for i, part := range parts {
			if part == "Up" && i+1 < len(parts) {
				return parts[i+1] + "m"
			}
		}
	}
	if strings.Contains(status, "hour") {
		parts := strings.Split(status, " ")
		for i, part := range parts {
			if part == "Up" && i+1 < len(parts) {
				return parts[i+1] + "h"
			}
		}
	}
	if strings.Contains(status, "day") {
		parts := strings.Split(status, " ")
		for i, part := range parts {
			if part == "Up" && i+1 < len(parts) {
				return parts[i+1] + "d"
			}
		}
	}

	return "running"
}

