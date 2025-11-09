package deploy

import (
	"fmt"
	"time"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/nginx"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ssl"
	"github.com/ekinertac/podlift/internal/ui"
)

// SetupLoadBalancer configures nginx load balancer across multiple servers
func SetupLoadBalancer(cfg *config.Config, version string) error {
	allServers := cfg.GetAllServers()
	
	// If only one server, no load balancing needed
	if len(allServers) <= 1 {
		return nil
	}

	// Find primary server (where nginx will run)
	primaryServer, _, err := cfg.GetPrimaryServer()
	if err != nil {
		return err
	}

	fmt.Println(ui.Info(fmt.Sprintf("Setting up load balancer on %s", primaryServer.Host)))
	fmt.Println(ui.Info(fmt.Sprintf("  Balancing across %d servers", len(allServers))))
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
		return fmt.Errorf("failed to connect to primary server: %w", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		return err
	}

	// Collect upstreams from all servers
	var upstreams []nginx.Upstream
	
	for _, srv := range allServers {
		for serviceName := range cfg.Services {
			// Load balancer proxies to nginx on each server (port 80)
			// nginx on each server then proxies to the local containers
			upstreams = append(upstreams, nginx.Upstream{
				Name: fmt.Sprintf("%s-%s", srv.Host, serviceName),
				Host: srv.Host,
				Port: 80, // Always use port 80 (nginx)
			})
		}
	}

	if len(upstreams) == 0 {
		return fmt.Errorf("no upstreams found")
	}

	fmt.Println(ui.Info(fmt.Sprintf("  Configuring %d upstreams", len(upstreams))))

	// Create nginx manager
	nginxMgr := nginx.NewManager(client)

	// Ensure nginx installed
	if installed, _ := nginxMgr.IsInstalled(); !installed {
		if err := nginxMgr.Install(); err != nil {
			return err
		}
	}

	// Check for SSL
	sslCfg := nginx.SSLConfig{Enabled: false}
	domain := cfg.Domain
	if domain == "" {
		domain = primaryServer.Host
	}

	if cfg.Proxy != nil && cfg.Proxy.SSL == "letsencrypt" {
		certbotMgr := ssl.NewCertbotManager(client)
		certExists, _ := certbotMgr.CheckCertificate(domain)
		
		if certExists {
			sslCfg = nginx.SSLConfig{
				Enabled:     true,
				CertPath:    certbotMgr.GetCertificatePath(domain),
				KeyPath:     certbotMgr.GetKeyPath(domain),
				LetsEncrypt: true,
			}
			fmt.Println(ui.Success("  Using SSL certificate"))
		}
	}

	// Update nginx configuration
	if err := nginxMgr.UpdateUpstream(cfg.Service, upstreams, domain, sslCfg); err != nil {
		return fmt.Errorf("failed to configure load balancer: %w", err)
	}

	fmt.Println(ui.Success("Load balancer configured"))
	fmt.Println()
	
	return nil
}

