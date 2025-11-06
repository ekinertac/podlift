package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCommand)
}

var initCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize podlift configuration",
	Long:  "Creates a podlift.yml configuration file and .env.example in the current directory",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(cwd, "podlift.yml")
	envExamplePath := filepath.Join(cwd, ".env.example")

	// Check if podlift.yml already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println(ui.Error("podlift.yml already exists"))
		fmt.Println()
		fmt.Println("To reinitialize, delete the existing file first:")
		fmt.Println(ui.Code("  rm podlift.yml"))
		return fmt.Errorf("podlift.yml already exists")
	}

	// Detect project characteristics (optional - try to be smart)
	projectName := detectProjectName(cwd)
	hasDockerfile := fileExists(filepath.Join(cwd, "Dockerfile"))

	// Generate config content
	configContent := generateConfig(projectName, hasDockerfile)

	// Write podlift.yml
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write podlift.yml: %w", err)
	}
	fmt.Println(ui.Success("Created podlift.yml"))

	// Generate .env.example content
	envContent := generateEnvExample()

	// Write .env.example
	if err := os.WriteFile(envExamplePath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env.example: %w", err)
	}
	fmt.Println(ui.Success("Created .env.example"))

	// Show next steps
	fmt.Println()
	fmt.Println(ui.Title("Next steps:"))
	fmt.Println(ui.Info("1. Edit podlift.yml and add your server"))
	fmt.Println(ui.Info("2. Copy .env.example to .env and set secrets:"))
	fmt.Println(ui.Code("     cp .env.example .env"))
	fmt.Println(ui.Info("3. Prepare your server:"))
	fmt.Println(ui.Code("     podlift setup"))
	fmt.Println(ui.Info("4. Validate configuration:"))
	fmt.Println(ui.Code("     podlift validate"))
	fmt.Println(ui.Info("5. Deploy:"))
	fmt.Println(ui.Code("     podlift deploy"))

	if !hasDockerfile {
		fmt.Println()
		fmt.Println(ui.Warning("No Dockerfile found"))
		fmt.Println("Create a Dockerfile in the project root before deploying.")
	}

	return nil
}

// detectProjectName tries to determine a good service name from the directory
func detectProjectName(dir string) string {
	name := filepath.Base(dir)
	
	// Sanitize name (remove invalid characters)
	// Service names should be lowercase alphanumeric with hyphens
	sanitized := ""
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			sanitized += string(ch)
		} else if ch >= 'A' && ch <= 'Z' {
			sanitized += string(ch + 32) // Convert to lowercase
		} else if ch == '_' || ch == ' ' {
			sanitized += "-"
		}
	}
	
	if sanitized == "" {
		return "myapp"
	}
	
	return sanitized
}

// generateConfig generates the podlift.yml content
func generateConfig(serviceName string, hasDockerfile bool) string {
	config := fmt.Sprintf(`service: %s
image: %s

# Add your server (required)
servers:
  - host: YOUR_SERVER_IP
    user: root
    ssh_key: ~/.ssh/id_rsa

# Uncomment to use a Docker registry
# registry:
#   server: ghcr.io
#   username: ${REGISTRY_USER}
#   password: ${REGISTRY_PASSWORD}

# Uncomment to add PostgreSQL
# dependencies:
#   postgres:
#     image: postgres:16
#     port: 5432
#     volume: postgres_data:/var/lib/postgresql/data
#     env:
#       POSTGRES_PASSWORD: ${DB_PASSWORD}
#       POSTGRES_DB: %s

# Web service configuration
services:
  web:
    port: 8000
    replicas: 2
    healthcheck:
      path: /health
      expect: [200]
    env:
      SECRET_KEY: ${SECRET_KEY}
      # DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/%s

# Uncomment to enable SSL with Let's Encrypt
# proxy:
#   enabled: true
#   ssl: letsencrypt
#   ssl_email: admin@example.com

# Uncomment to add deployment hooks
# hooks:
#   after_deploy:
#     - docker exec %s-web-1 echo "Deployment complete"
`, serviceName, serviceName, serviceName, serviceName, serviceName)

	if !hasDockerfile {
		config = "# WARNING: No Dockerfile found in current directory\n" +
			"# Create a Dockerfile before deploying\n\n" + config
	}

	return config
}

// generateEnvExample generates the .env.example content
func generateEnvExample() string {
	return `# Environment variables for podlift
# Copy to .env and fill in your values
# WARNING: Never commit .env to git!

# Docker Registry (if using registry instead of SCP)
# For GitHub: create token at https://github.com/settings/tokens
# REGISTRY_USER=your-github-username
# REGISTRY_PASSWORD=ghp_your_github_token

# Database (if using PostgreSQL)
# DB_PASSWORD=your-secure-database-password

# Application secrets
SECRET_KEY=your-application-secret-key

# Add your app-specific environment variables below
`
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

