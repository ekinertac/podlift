package commands

import (
	"fmt"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

var (
	execReplica int
)

var execCmd = &cobra.Command{
	Use:   "exec <service> <command>",
	Short: "Execute command in a running container",
	Long:  `Execute a command in a running container for the specified service.`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  runExec,
}

func init() {
	execCmd.Flags().IntVar(&execReplica, "replica", 1, "Execute on specific replica")
	rootCmd.AddCommand(execCmd)
}

func runExec(cmd *cobra.Command, args []string) error {
	serviceName := args[0]
	command := args[1:]

	// Load config
	cfg, err := config.Load("podlift.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get primary server
	primaryServer, _, err := cfg.GetPrimaryServer()
	if err != nil {
		return err
	}

	fmt.Println(ui.Title(fmt.Sprintf("Executing command on %s (replica %d)", serviceName, execReplica)))
	fmt.Println()

	// Connect to primary server
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

	// Build container name
	containerName := fmt.Sprintf("%s-%s-%d", cfg.Service, serviceName, execReplica)

	// Check if container exists
	checkCmd := fmt.Sprintf("sudo docker ps --filter name=%s --format '{{.Names}}'", containerName)
	output, err := client.Execute(checkCmd)
	if err != nil || output == "" {
		return fmt.Errorf("container %s not found or not running", containerName)
	}

	// Execute command in container
	execCommand := fmt.Sprintf("sudo docker exec -it %s %s", containerName, joinArgs(command))
	
	fmt.Println(ui.Info(fmt.Sprintf("Executing: %s", joinArgs(command))))
	fmt.Println()

	result, err := client.Execute(execCommand)
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	fmt.Println(result)
	return nil
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		// Quote args with spaces
		if containsSpace(arg) {
			result += fmt.Sprintf(`"%s"`, arg)
		} else {
			result += arg
		}
	}
	return result
}

func containsSpace(s string) bool {
	for _, r := range s {
		if r == ' ' {
			return true
		}
	}
	return false
}

