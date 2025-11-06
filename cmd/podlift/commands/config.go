package commands

import (
	"fmt"
	"strings"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configShowSecrets bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long:  `Display the current podlift configuration with environment variables resolved.`,
	RunE:  runConfig,
}

func init() {
	configCmd.Flags().BoolVar(&configShowSecrets, "show-secrets", false, "Include environment variable values")
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load("podlift.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.Title("Current Configuration"))
	fmt.Println()

	// Convert config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configStr := string(data)

	// Mask secrets if not showing them
	if !configShowSecrets {
		configStr = maskSecrets(configStr)
	}

	fmt.Println(configStr)

	return nil
}

func maskSecrets(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	secretKeys := []string{"password", "secret", "key", "token", "credentials"}
	
	for i, line := range lines {
		lower := strings.ToLower(line)
		for _, key := range secretKeys {
			if strings.Contains(lower, key+":") && !strings.Contains(line, "${") {
				// Mask the value
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					lines[i] = parts[0] + ": ********"
				}
			}
		}
	}
	
	return strings.Join(lines, "\n")
}

