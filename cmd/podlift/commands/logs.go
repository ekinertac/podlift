package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/docker"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   int
	logsSince  string
)

func init() {
	rootCmd.AddCommand(logsCommand)
	logsCommand.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCommand.Flags().IntVarP(&logsTail, "tail", "n", 100, "Number of lines to show from the end of the logs")
	logsCommand.Flags().StringVar(&logsSince, "since", "", "Show logs since timestamp (e.g. 2h, 30m)")
}

var logsCommand = &cobra.Command{
	Use:   "logs <service>",
	Short: "View container logs",
	Long:  "Display logs from a service container",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runLogs,
}

func runLogs(cmd *cobra.Command, args []string) error {
	serviceName := args[0] // e.g., "web", "worker"

	// Load configuration
	configPath, err := config.Find()
	if err != nil {
		return fmt.Errorf("podlift.yml not found")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Get primary server (show logs from first server)
	primaryServer, _, err := cfg.GetPrimaryServer()
	if err != nil {
		return err
	}

	// Create SSH client
	sshClient, err := ssh.NewClient(ssh.Config{
		Host:    primaryServer.Host,
		Port:    primaryServer.Port,
		User:    primaryServer.User,
		KeyPath: primaryServer.SSHKey,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return err
	}
	defer sshClient.Close()

	// Find container name (get first container for this service)
	findCmd := fmt.Sprintf(`docker ps --filter "label=podlift.service=%s" --filter "label=podlift.container_type=%s" --format "{{.Names}}" | head -1`, cfg.Service, serviceName)
	
	containerName, err := sshClient.Execute(findCmd)
	if err != nil || strings.TrimSpace(containerName) == "" {
		// Try without container_type filter (for now)
		findCmd = fmt.Sprintf(`docker ps --filter "label=podlift.service=%s" --format "{{.Names}}" | grep "-%s-" | head -1`, cfg.Service, serviceName)
		containerName, err = sshClient.Execute(findCmd)
		if err != nil || strings.TrimSpace(containerName) == "" {
			return fmt.Errorf("no running container found for service '%s'", serviceName)
		}
	}

	containerName = strings.TrimSpace(containerName)

	fmt.Println(ui.Info(fmt.Sprintf("Logs for %s (%s)", serviceName, containerName)))
	fmt.Println()

	// Build logs command
	logsCmd := docker.GenerateLogsCommand(containerName, logsTail, logsFollow)
	
	if logsSince != "" {
		logsCmd += fmt.Sprintf(" --since %s", logsSince)
	}

	// Stream logs
	if logsFollow {
		// For follow mode, stream indefinitely
		return sshClient.ExecuteWithOutput(logsCmd, os.Stdout, os.Stderr)
	}

	// For non-follow mode, just show output
	output, err := sshClient.Execute(logsCmd)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	fmt.Print(output)

	return nil
}

