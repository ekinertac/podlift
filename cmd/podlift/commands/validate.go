package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/git"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(validateCommand)
}

var validateCommand = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and verify server readiness",
	Long:  "Performs pre-flight checks before deployment",
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	fmt.Println(ui.Title("Validating configuration..."))
	fmt.Println()

	// 1. Find and load config
	configPath, err := config.Find()
	if err != nil {
		fmt.Println(ui.Error("podlift.yml not found"))
		fmt.Println()
		fmt.Println("Run: podlift init")
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Println(ui.Error("Configuration invalid"))
		fmt.Println()
		fmt.Println(err.Error())
		return err
	}
	fmt.Println(ui.Success("Configuration valid"))

	// 2. Load and substitute environment variables
	if err := cfg.SubstituteConfigEnvVars(); err != nil {
		fmt.Println(ui.Error("Failed to load environment variables"))
		return err
	}

	// 3. Check git state
	if git.IsRepository() {
		if err := git.RequireCleanState(); err != nil {
			fmt.Println(ui.Error("Git working tree is dirty"))
			fmt.Println()
			fmt.Println(err.Error())
			return err
		}
		
		gitInfo, _ := git.GetInfo()
		fmt.Println(ui.Success("Git state clean"))
		if gitInfo != "" {
			for _, line := range splitLines(gitInfo) {
				fmt.Println(ui.Info("  " + line))
			}
		}
	} else {
		fmt.Println(ui.Warning("Not a git repository"))
		fmt.Println(ui.Info("  Git is recommended for version tracking"))
	}

	// 4. Validate servers
	fmt.Println()
	allServers := cfg.GetAllServers()
	fmt.Printf("Checking %d server(s)...\n", len(allServers))
	fmt.Println()

	for i, serverWithRole := range allServers {
		server := serverWithRole.Server
		fmt.Printf("[%d/%d] %s (%s role)\n", i+1, len(allServers), server.Host, serverWithRole.Role)

		// Create SSH client
		sshClient, err := ssh.NewClient(ssh.Config{
			Host:    server.Host,
			Port:    server.Port,
			User:    server.User,
			KeyPath: server.SSHKey,
			Timeout: 10 * time.Second,
		})
		if err != nil {
			fmt.Println(ui.Error(fmt.Sprintf("  Failed to create SSH client: %v", err)))
			return err
		}
		defer sshClient.Close()

		// Test SSH connection
		if err := sshClient.TestConnection(); err != nil {
			fmt.Println(ui.Error("  SSH connection failed"))
			fmt.Println()
			fmt.Println(fmt.Sprintf("Cannot connect to %s@%s:%d", server.User, server.Host, server.Port))
			fmt.Println()
			fmt.Println("Ensure SSH key authentication is configured.")
			return err
		}
		fmt.Println(ui.Success("  SSH connection"))

		// Check Docker
		dockerVersion, err := sshClient.CheckDocker()
		if err != nil {
			fmt.Println(ui.Error("  Docker not installed"))
			fmt.Println()
			fmt.Println("Install Docker: podlift setup")
			return fmt.Errorf("Docker not installed on %s", server.Host)
		}
		fmt.Println(ui.Success(fmt.Sprintf("  Docker %s", trimVersion(dockerVersion))))

		// Check disk space
		diskSpace, err := sshClient.GetDiskSpace()
		if err == nil {
			if diskSpace < 10.0 {
				fmt.Println(ui.Warning(fmt.Sprintf("  Low disk space: %.1fGB available", diskSpace)))
			} else {
				fmt.Println(ui.Success(fmt.Sprintf("  Disk space: %.1fGB available", diskSpace)))
			}
		}

		// Check for service conflicts
		exists, info, _ := sshClient.CheckExistingService(cfg.Service)
		if exists {
			// Get current git repo
			currentRepo, _ := git.GetRemoteURL()
			sameRepo, _ := sshClient.CompareGitRepo(cfg.Service, currentRepo)
			
			if sameRepo {
				fmt.Println(ui.Info(fmt.Sprintf("  Service '%s' exists (redeployment)", cfg.Service)))
				fmt.Println(ui.Info(fmt.Sprintf("    Version: %s", info.Version)))
			} else {
				fmt.Println(ui.Warning(fmt.Sprintf("  Service '%s' exists with different repository", cfg.Service)))
				fmt.Println()
				fmt.Println("This may be a different application!")
				fmt.Println("Use a unique service name to avoid conflicts.")
			}
		}

		fmt.Println()
	}

	// 5. Summary
	fmt.Println(ui.Success("All checks passed!"))
	fmt.Println()
	fmt.Println(ui.Info("Ready to deploy: podlift deploy"))

	return nil
}

// Helper functions
func trimVersion(version string) string {
	// Extract just "Docker version 24.0.5" from full output
	parts := splitLines(version)
	if len(parts) > 0 {
		return parts[0]
	}
	return version
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

