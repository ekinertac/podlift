package hooks

import (
	"fmt"
	"strings"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/ssh"
	"github.com/ekinertac/podlift/internal/ui"
)

// Execute runs deployment hooks on the primary server
func Execute(client ssh.SSHClient, cfg *config.Config, stage string) error {
	if cfg.Hooks == nil {
		return nil
	}

	var commands []string
	switch stage {
	case "before_deploy":
		commands = cfg.Hooks.BeforeDeploy
	case "after_deploy":
		commands = cfg.Hooks.AfterDeploy
	case "after_rollback":
		commands = cfg.Hooks.AfterRollback
	default:
		return fmt.Errorf("unknown hook stage: %s", stage)
	}

	if len(commands) == 0 {
		return nil
	}

	fmt.Println(ui.Info(fmt.Sprintf("Running %s hooks...", stage)))

	for i, cmd := range commands {
		fmt.Println(ui.Info(fmt.Sprintf("  [%d/%d] %s", i+1, len(commands), cmd)))
		
		output, err := client.Execute(cmd)
		if err != nil {
			return fmt.Errorf("hook failed: %w", err)
		}

		if output != "" {
			// Show command output
			lines := strings.Split(strings.TrimSpace(output), "\n")
			for _, line := range lines {
				fmt.Println(fmt.Sprintf("    %s", line))
			}
		}
	}

	fmt.Println(ui.Success("  ✓ Hooks completed"))
	fmt.Println()

	return nil
}

