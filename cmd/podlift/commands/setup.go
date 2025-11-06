package commands

import (
	"fmt"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/setup"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

var (
	setupNoFirewall bool
	setupNoSecurity bool
)

func init() {
	rootCmd.AddCommand(setupCommand)
	setupCommand.Flags().BoolVar(&setupNoFirewall, "no-firewall", false, "Skip firewall configuration")
	setupCommand.Flags().BoolVar(&setupNoSecurity, "no-security", false, "Skip security hardening")
}

var setupCommand = &cobra.Command{
	Use:   "setup",
	Short: "Prepare servers for deployment",
	Long:  "Installs Docker, configures firewall, and applies basic security hardening",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Load configuration
	configPath, err := config.Find()
	if err != nil {
		return fmt.Errorf("podlift.yml not found. Run: podlift init")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Load environment variables
	if err := cfg.SubstituteConfigEnvVars(); err != nil {
		return err
	}

	fmt.Println(ui.Title("Setting up servers..."))
	fmt.Println()

	// Setup each server
	allServers := cfg.GetAllServers()
	for i, serverWithRole := range allServers {
		server := serverWithRole.Server
		
		fmt.Printf("[%d/%d] Server: %s (%s)\n", i+1, len(allServers), server.Host, serverWithRole.Role)
		fmt.Println()

		// Create SSH client
		sshClient, err := ssh.NewClient(ssh.Config{
			Host:    server.Host,
			Port:    server.Port,
			User:    server.User,
			KeyPath: server.SSHKey,
			Timeout: 30 * time.Second,
		})
		if err != nil {
			return fmt.Errorf("failed to create SSH client: %w", err)
		}
		defer sshClient.Close()

		// Connect
		fmt.Println(ui.Info("[1/4] Connecting to server..."))
		if err := sshClient.Connect(); err != nil {
			return fmt.Errorf("failed to connect to %s: %w", server.Host, err)
		}
		fmt.Println(ui.Success("Connected"))
		fmt.Println()

		// Install Docker
		fmt.Println(ui.Info("[2/4] Installing Docker..."))
		if err := setup.InstallDocker(sshClient); err != nil {
			return fmt.Errorf("Docker installation failed on %s: %w", server.Host, err)
		}
		fmt.Println()

		// Configure firewall
		if !setupNoFirewall {
			fmt.Println(ui.Info("[3/4] Configuring firewall..."))
			if err := setup.ConfigureFirewall(sshClient); err != nil {
				fmt.Println(ui.Warning(fmt.Sprintf("Firewall configuration failed: %v", err)))
				fmt.Println(ui.Info("Skip firewall: podlift setup --no-firewall"))
			}
			fmt.Println()
		} else {
			fmt.Println(ui.Info("[3/4] Skipping firewall configuration (--no-firewall)"))
			fmt.Println()
		}

		// Apply security
		if !setupNoSecurity {
			fmt.Println(ui.Info("[4/4] Applying security settings..."))
			if err := setup.ApplySecurity(sshClient); err != nil {
				fmt.Println(ui.Warning(fmt.Sprintf("Security configuration failed: %v", err)))
				fmt.Println(ui.Info("Skip security: podlift setup --no-security"))
			}
			fmt.Println()
		} else {
			fmt.Println(ui.Info("[4/4] Skipping security hardening (--no-security)"))
			fmt.Println()
		}

		// Verify setup
		fmt.Println(ui.Info("Verifying setup..."))
		if err := setup.VerifySetup(sshClient); err != nil {
			fmt.Println(ui.Warning(fmt.Sprintf("Verification warning: %v", err)))
		}
		fmt.Println()

		fmt.Println(ui.Success(fmt.Sprintf("Server %s setup complete!", server.Host)))
		fmt.Println()
	}

	fmt.Println(ui.Title("Setup complete!"))
	fmt.Println()
	fmt.Println(ui.Info("Next step: podlift deploy"))

	return nil
}

