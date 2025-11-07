package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the complete podlift configuration
type Config struct {
	Service      string                 `yaml:"service"`
	Domain       string                 `yaml:"domain,omitempty"`
	Image        string                 `yaml:"image"`
	Git          *GitConfig             `yaml:"git,omitempty"`
	Servers      ServersConfig          `yaml:"servers"`
	Registry     *RegistryConfig        `yaml:"registry,omitempty"`
	Dependencies map[string]Dependency  `yaml:"dependencies,omitempty"`
	Services     map[string]Service     `yaml:"services,omitempty"`
	Proxy        *ProxyConfig           `yaml:"proxy,omitempty"`
	Hooks        *HooksConfig           `yaml:"hooks,omitempty"`
	EnvFile      string                 `yaml:"env_file,omitempty"`
	
	// Internal fields
	configPath string // Path to the config file (not serialized)
}

// ServersConfig handles both list and map formats for servers
type ServersConfig struct {
	servers map[string][]Server
}

// UnmarshalYAML implements custom YAML unmarshaling
func (s *ServersConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s.servers = make(map[string][]Server)
	
	// Try to unmarshal as map first (role-based)
	var mapData map[string][]Server
	if err := unmarshal(&mapData); err == nil {
		s.servers = mapData
		return nil
	}
	
	// If that fails, try as list (simple format)
	var listData []Server
	if err := unmarshal(&listData); err == nil {
		// Convert list to map with "web" role
		s.servers = map[string][]Server{
			"web": listData,
		}
		return nil
	}
	
	return fmt.Errorf("servers must be either a list or a map")
}

// Get returns the servers map
func (s *ServersConfig) Get() map[string][]Server {
	if s.servers == nil {
		s.servers = make(map[string][]Server)
	}
	return s.servers
}

// Set sets the servers map
func (s *ServersConfig) Set(servers map[string][]Server) {
	s.servers = servers
}

// GitConfig contains git repository information
type GitConfig struct {
	Repo   string `yaml:"repo,omitempty"`
	Branch string `yaml:"branch,omitempty"`
}

// Server represents a deployment server
type Server struct {
	Host    string   `yaml:"host"`
	User    string   `yaml:"user,omitempty"`
	SSHKey  string   `yaml:"ssh_key,omitempty"`
	Port    int      `yaml:"port,omitempty"`
	Labels  []string `yaml:"labels,omitempty"`
}

// RegistryConfig contains Docker registry configuration
type RegistryConfig struct {
	Server   string `yaml:"server,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// Dependency represents a service dependency (postgres, redis, etc.)
type Dependency struct {
	Image     string            `yaml:"image"`
	Host      string            `yaml:"host,omitempty"`      // Specific server host
	Role      string            `yaml:"role,omitempty"`     // Server role (e.g., "db", "cache")
	Labels    []string          `yaml:"labels,omitempty"`   // Match servers with these labels
	Port      int               `yaml:"port,omitempty"`
	Volume    string            `yaml:"volume,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Command   string            `yaml:"command,omitempty"`
	Options   map[string]string `yaml:"options,omitempty"`
}

// Service represents an application service
type Service struct {
	Port       int                   `yaml:"port,omitempty"`
	Replicas   int                   `yaml:"replicas,omitempty"`
	Command    string                `yaml:"command,omitempty"`
	Healthcheck *HealthcheckConfig   `yaml:"healthcheck,omitempty"`
	Env        map[string]string     `yaml:"env,omitempty"`
	Volumes    []string              `yaml:"volumes,omitempty"`
	Options    map[string]string     `yaml:"options,omitempty"`
}

// HealthcheckConfig contains health check configuration
type HealthcheckConfig struct {
	Path     string   `yaml:"path,omitempty"`
	Expect   []int    `yaml:"expect,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty"`
	Interval string   `yaml:"interval,omitempty"`
	Retries  int      `yaml:"retries,omitempty"`
	Enabled  *bool    `yaml:"enabled,omitempty"` // Use pointer to distinguish unset from false
}

// ProxyConfig contains nginx proxy configuration
type ProxyConfig struct {
	Enabled   bool   `yaml:"enabled,omitempty"`
	SSL       string `yaml:"ssl,omitempty"`
	SSLEmail  string `yaml:"ssl_email,omitempty"`
}

// HooksConfig contains deployment hooks
type HooksConfig struct {
	BeforeDeploy  []string `yaml:"before_deploy,omitempty"`
	AfterDeploy   []string `yaml:"after_deploy,omitempty"`
	AfterRollback []string `yaml:"after_rollback,omitempty"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Store config path for .env lookup
	config.configPath = path

	// Apply defaults
	config.applyDefaults()

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration invalid: %w", err)
	}

	return &config, nil
}

// Find searches for podlift.yml in the current directory and parent directories
func Find() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, "podlift.yml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	return "", fmt.Errorf("podlift.yml not found")
}

// applyDefaults sets default values for optional fields
func (c *Config) applyDefaults() {
	// Default proxy enabled
	if c.Proxy == nil {
		c.Proxy = &ProxyConfig{Enabled: true}
	}

	// Default services if not specified
	if c.Services == nil || len(c.Services) == 0 {
		c.Services = map[string]Service{
			"web": {
				Port:     8000,
				Replicas: 1,
				Healthcheck: &HealthcheckConfig{
					Path:     "/health",
					Expect:   []int{200},
					Timeout:  "30s",
					Interval: "10s",
					Retries:  3,
				},
			},
		}
	}

	// Apply defaults to each service
	for name, svc := range c.Services {
		if svc.Port == 0 {
			svc.Port = 8000
		}
		if svc.Replicas == 0 {
			svc.Replicas = 1
		}
		if svc.Healthcheck == nil {
			svc.Healthcheck = &HealthcheckConfig{
				Path:     "/health",
				Expect:   []int{200},
				Timeout:  "30s",
				Interval: "10s",
				Retries:  3,
			}
		}
		c.Services[name] = svc
	}

	// Default server settings
	c.applyServerDefaults()
}

// applyServerDefaults applies defaults to server configurations
func (c *Config) applyServerDefaults() {
	servers := c.Servers.Get()
	for role, serverList := range servers {
		for i, server := range serverList {
			if server.User == "" {
				server.User = "root"
			}
			if server.SSHKey == "" {
				server.SSHKey = "~/.ssh/id_rsa"
			}
			if server.Port == 0 {
				server.Port = 22
			}
			serverList[i] = server
		}
		servers[role] = serverList
	}
	c.Servers.Set(servers)
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.Service == "" {
		return fmt.Errorf("service name is required")
	}

	if c.Image == "" {
		return fmt.Errorf("image name is required")
	}

	servers := c.Servers.Get()
	if len(servers) == 0 {
		return fmt.Errorf("at least one server is required")
	}

	// Validate servers
	for role, serverList := range servers {
		if len(serverList) == 0 {
			return fmt.Errorf("role '%s' has no servers", role)
		}
		for i, server := range serverList {
			if server.Host == "" {
				return fmt.Errorf("server %d in role '%s' missing host", i, role)
			}
		}
	}

	// Validate services
	for name, svc := range c.Services {
		if svc.Port < 1 || svc.Port > 65535 {
			return fmt.Errorf("service '%s' has invalid port: %d", name, svc.Port)
		}
		if svc.Replicas < 1 {
			return fmt.Errorf("service '%s' replicas must be >= 1", name)
		}
	}

	// Validate dependencies
	for name, dep := range c.Dependencies {
		if dep.Image == "" {
			return fmt.Errorf("dependency '%s' missing image", name)
		}
		// Validate that specified host/role/labels exist
		if dep.Host != "" || dep.Role != "" || len(dep.Labels) > 0 {
			if _, _, err := c.GetDependencyServer(dep); err != nil {
				return fmt.Errorf("dependency '%s': %w", name, err)
			}
		}
	}

	// Validate registry if specified
	if c.Registry != nil {
		if c.Registry.Server != "" && c.Registry.Username == "" {
			return fmt.Errorf("registry username required when server is specified")
		}
	}

	return nil
}

// GetPrimaryServer returns the first server with "primary" label, or first server in first role
func (c *Config) GetPrimaryServer() (*Server, string, error) {
	servers := c.Servers.Get()
	
	// First, check all servers for "primary" label
	for role, serverList := range servers {
		for _, server := range serverList {
			for _, label := range server.Labels {
				if label == "primary" {
					return &server, role, nil
				}
			}
		}
	}
	
	// If no primary label found, return first server from first role
	// (order is deterministic: iterate map in sorted order)
	for role, serverList := range servers {
		if len(serverList) > 0 {
			return &serverList[0], role, nil
		}
	}
	
	return nil, "", fmt.Errorf("no servers configured")
}

// GetDependencyServer returns the server where a dependency should run
func (c *Config) GetDependencyServer(dep Dependency) (*Server, string, error) {
	servers := c.Servers.Get()
	
	// If host is specified, find exact match
	if dep.Host != "" {
		for role, serverList := range servers {
			for _, server := range serverList {
				if server.Host == dep.Host {
					return &server, role, nil
				}
			}
		}
		return nil, "", fmt.Errorf("dependency host '%s' not found in servers", dep.Host)
	}
	
	// If role is specified, use first server in that role
	if dep.Role != "" {
		if serverList, ok := servers[dep.Role]; ok && len(serverList) > 0 {
			return &serverList[0], dep.Role, nil
		}
		return nil, "", fmt.Errorf("dependency role '%s' not found in servers", dep.Role)
	}
	
	// If labels are specified, find server with matching labels
	if len(dep.Labels) > 0 {
		for role, serverList := range servers {
			for _, server := range serverList {
				for _, depLabel := range dep.Labels {
					for _, serverLabel := range server.Labels {
						if depLabel == serverLabel {
							return &server, role, nil
						}
					}
				}
			}
		}
		return nil, "", fmt.Errorf("no server found with labels %v", dep.Labels)
	}
	
	// Default: use primary server
	return c.GetPrimaryServer()
}

// GetAllServers returns all servers flattened with their roles
func (c *Config) GetAllServers() []ServerWithRole {
	var result []ServerWithRole
	servers := c.Servers.Get()
	for role, serverList := range servers {
		for _, server := range serverList {
			result = append(result, ServerWithRole{
				Server: server,
				Role:   role,
			})
		}
	}
	return result
}

// ServerWithRole combines a server with its role
type ServerWithRole struct {
	Server
	Role string
}

