package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadEnv loads environment variables from .env file
func LoadEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // .env file is optional
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		os.Setenv(key, value)
	}

	return nil
}

// SubstituteEnvVars replaces ${VAR} patterns with environment variables
func SubstituteEnvVars(value string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		
		// Handle default value syntax: ${VAR:-default}
		if strings.Contains(varName, ":-") {
			parts := strings.SplitN(varName, ":-", 2)
			varName = parts[0]
			defaultValue := parts[1]
			if val := os.Getenv(varName); val != "" {
				return val
			}
			return defaultValue
		}

		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // Return original if not found
	})
}

// ExpandPath expands ~ and environment variables in paths
func ExpandPath(path string) string {
	// Expand ~
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = strings.Replace(path, "~", home, 1)
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path
}

// FindEnvFile looks for .env in the same directory as podlift.yml
func FindEnvFile(configPath string) string {
	if configPath == "" {
		return ""
	}
	
	// Get directory containing config file
	configDir := filepath.Dir(configPath)
	envPath := filepath.Join(configDir, ".env")
	
	// Check if .env exists
	if _, err := os.Stat(envPath); err == nil {
		return envPath
	}
	
	return ""
}

// SubstituteConfigEnvVars substitutes environment variables in config
func (c *Config) SubstituteConfigEnvVars() error {
	// Load .env file if it exists (same directory as podlift.yml)
	envPath := FindEnvFile(c.configPath)
	if envPath != "" {
		if err := LoadEnv(envPath); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// Substitute in registry
	if c.Registry != nil {
		c.Registry.Username = SubstituteEnvVars(c.Registry.Username)
		c.Registry.Password = SubstituteEnvVars(c.Registry.Password)
	}

	// Substitute in servers (SSH key paths)
	servers := c.Servers.Get()
	for role, serverList := range servers {
		for i, server := range serverList {
			server.SSHKey = ExpandPath(SubstituteEnvVars(server.SSHKey))
			serverList[i] = server
		}
		servers[role] = serverList
	}
	c.Servers.Set(servers)

	// Substitute in dependencies
	for name, dep := range c.Dependencies {
		for key, value := range dep.Env {
			dep.Env[key] = SubstituteEnvVars(value)
		}
		c.Dependencies[name] = dep
	}

	// Substitute in services
	for name, svc := range c.Services {
		for key, value := range svc.Env {
			svc.Env[key] = SubstituteEnvVars(value)
		}
		c.Services[name] = svc
	}

	return nil
}

