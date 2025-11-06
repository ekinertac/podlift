package commands

import (
	"fmt"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/nginx"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ssl"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sslCommand)
	sslCommand.AddCommand(sslSetupCommand)
	sslCommand.AddCommand(sslRenewCommand)
	sslCommand.AddCommand(sslStatusCommand)
}

var sslCommand = &cobra.Command{
	Use:   "ssl",
	Short: "Manage SSL certificates",
	Long:  "Manage SSL/TLS certificates with Let's Encrypt",
}

var sslSetupCommand = &cobra.Command{
	Use:   "setup",
	Short: "Setup SSL certificates",
	Long:  "Obtain SSL certificates from Let's Encrypt and configure nginx",
	RunE:  runSSLSetup,
}

var sslRenewCommand = &cobra.Command{
	Use:   "renew",
	Short: "Renew SSL certificates",
	Long:  "Manually renew all SSL certificates",
	RunE:  runSSLRenew,
}

var sslStatusCommand = &cobra.Command{
	Use:   "status",
	Short: "Show SSL certificate status",
	Long:  "Display information about SSL certificates",
	RunE:  runSSLStatus,
}

func runSSLSetup(cmd *cobra.Command, args []string) error {
	// Load configuration
	configPath, err := config.Find()
	if err != nil {
		return fmt.Errorf("podlift.yml not found")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Check if SSL is configured
	if cfg.Proxy == nil || cfg.Proxy.SSL == "" || cfg.Proxy.SSL == "false" {
		fmt.Println(ui.Error("SSL not configured"))
		fmt.Println()
		fmt.Println("Add to podlift.yml:")
		fmt.Println(ui.Code(`proxy:
  enabled: true
  ssl: letsencrypt
  ssl_email: admin@example.com`))
		return fmt.Errorf("SSL not configured")
	}

	if cfg.Domain == "" {
		return fmt.Errorf("domain not configured in podlift.yml")
	}

	email := cfg.Proxy.SSLEmail
	if email == "" {
		return fmt.Errorf("ssl_email not configured in podlift.yml")
	}

	fmt.Println(ui.Title(fmt.Sprintf("Setting up SSL for %s", cfg.Domain)))
	fmt.Println()

	// Get primary server
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
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return err
	}
	defer sshClient.Close()

	// Create managers
	certbotMgr := ssl.NewCertbotManager(sshClient)
	nginxMgr := nginx.NewManager(sshClient)

	// 1. Ensure nginx is installed
	fmt.Println(ui.Info("[1/5] Checking nginx..."))
	if installed, _ := nginxMgr.IsInstalled(); !installed {
		if err := nginxMgr.Install(); err != nil {
			return err
		}
	} else {
		fmt.Println(ui.Success("nginx installed"))
	}

	// 2. Install certbot
	fmt.Println(ui.Info("[2/5] Installing certbot..."))
	if err := certbotMgr.Install(); err != nil {
		return err
	}

	// 3. Setup webroot for verification
	fmt.Println(ui.Info("[3/5] Setting up webroot..."))
	sshClient.Execute("sudo mkdir -p /var/www/html")
	sshClient.Execute("sudo chown -R www-data:www-data /var/www/html")

	// 4. Obtain certificate
	fmt.Println(ui.Info("[4/5] Obtaining SSL certificate..."))
	fmt.Println(ui.Info(fmt.Sprintf("  Domain: %s", cfg.Domain)))
	fmt.Println(ui.Info(fmt.Sprintf("  Email: %s", email)))
	
	if err := certbotMgr.ObtainCertificate(cfg.Domain, email); err != nil {
		return err
	}

	// 5. Setup auto-renewal
	fmt.Println(ui.Info("[5/5] Configuring auto-renewal..."))
	if err := certbotMgr.SetupAutoRenewal(); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(ui.Success("SSL setup complete!"))
	fmt.Println()
	fmt.Println(ui.Info("Certificate will auto-renew"))
	fmt.Println(ui.Info(fmt.Sprintf("Next: podlift deploy (will use HTTPS)")))

	return nil
}

func runSSLRenew(cmd *cobra.Command, args []string) error {
	// Load config to get server
	configPath, err := config.Find()
	if err != nil {
		return fmt.Errorf("podlift.yml not found")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

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
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return err
	}
	defer sshClient.Close()

	certbotMgr := ssl.NewCertbotManager(sshClient)
	
	return certbotMgr.RenewCertificates()
}

func runSSLStatus(cmd *cobra.Command, args []string) error {
	// Load config
	configPath, err := config.Find()
	if err != nil {
		return fmt.Errorf("podlift.yml not found")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if cfg.Domain == "" {
		return fmt.Errorf("domain not configured")
	}

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
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return err
	}
	defer sshClient.Close()

	certbotMgr := ssl.NewCertbotManager(sshClient)

	// Check if certificate exists
	exists, err := certbotMgr.CheckCertificate(cfg.Domain)
	if err != nil {
		return err
	}

	fmt.Println(ui.Title(fmt.Sprintf("SSL Status for %s", cfg.Domain)))
	fmt.Println()

	if !exists {
		fmt.Println(ui.Warning("No SSL certificate found"))
		fmt.Println()
		fmt.Println("Run: podlift ssl setup")
		return nil
	}

	fmt.Println(ui.Success("SSL certificate installed"))
	fmt.Println()

	// Get certificate info
	info, err := certbotMgr.GetCertificateInfo(cfg.Domain)
	if err == nil {
		for key, value := range info {
			fmt.Println(ui.Info(fmt.Sprintf("%s: %s", key, value)))
		}
	}

	return nil
}

