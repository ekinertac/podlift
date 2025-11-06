package commands

import (
	"fmt"

	"github.com/ekinertac/podlift/internal/config"
	"github.com/ekinertac/podlift/internal/deploy"
	"github.com/ekinertac/podlift/internal/git"
	"github.com/ekinertac/podlift/internal/ui"
	"github.com/spf13/cobra"
)

var (
	deploySkipBuild    bool
	deploySkipHealth   bool
	deployParallel     bool
	deployDryRun       bool
	deployZeroDowntime bool
)

func init() {
	rootCmd.AddCommand(deployCommand)
	deployCommand.Flags().BoolVar(&deploySkipBuild, "skip-build", false, "Skip building Docker image")
	deployCommand.Flags().BoolVar(&deploySkipHealth, "skip-healthcheck", false, "Skip health check")
	deployCommand.Flags().BoolVar(&deployParallel, "parallel", false, "Deploy to all servers in parallel")
	deployCommand.Flags().BoolVar(&deployDryRun, "dry-run", false, "Show what would happen without executing")
	deployCommand.Flags().BoolVar(&deployZeroDowntime, "zero-downtime", true, "Use zero-downtime deployment with nginx (default: true)")
}

var deployCommand = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your application",
	Long:  "Build and deploy your application to all configured servers",
	RunE:  runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// Load configuration
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

	// Load environment variables
	if err := cfg.SubstituteConfigEnvVars(); err != nil {
		return err
	}

	// Check git state (unless dry run)
	if !deployDryRun {
		if git.IsRepository() {
			if err := git.RequireCleanState(); err != nil {
				fmt.Println(ui.Error("Git working tree is dirty"))
				fmt.Println()
				fmt.Println(err.Error())
				return err
			}
		} else {
			fmt.Println(ui.Warning("Not a git repository"))
			fmt.Println(ui.Info("Git is recommended for version tracking"))
			fmt.Println()
		}
	}

	// Zero-downtime is default when proxy is enabled
	// Can be disabled with --no-zero-downtime flag (not implemented - why would you?)
	useZeroDowntime := deployZeroDowntime || (cfg.Proxy != nil && cfg.Proxy.Enabled)
	
	// Actually, just make it default always
	// Basic deployment only if explicitly requested
	if !deployZeroDowntime && cfg.Proxy == nil {
		// No proxy configured - basic deployment
		useZeroDowntime = false
	} else {
		// Default to zero-downtime
		useZeroDowntime = true
		fmt.Println(ui.Info("Using zero-downtime deployment"))
	}

	// Deploy
	deployOpts := deploy.DeployOptions{
		Config:       cfg,
		SkipBuild:    deploySkipBuild,
		SkipHealth:   deploySkipHealth,
		Parallel:     deployParallel,
		DryRun:       deployDryRun,
		ZeroDowntime: useZeroDowntime,
	}

	if err := deploy.Deploy(deployOpts); err != nil {
		fmt.Println()
		fmt.Println(ui.Error("Deployment failed"))
		fmt.Println()
		fmt.Println(err.Error())
		return err
	}

	return nil
}

