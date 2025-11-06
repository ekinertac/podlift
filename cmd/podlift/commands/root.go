package commands

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "podlift",
	Short: "Simple, transparent deployment for containerized applications",
	Long: `podlift deploys your Docker containers to any server with SSH access.
No black boxes, no magic, no broken promises.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show podlift version",
	Run: func(cmd *cobra.Command, args []string) {
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")) // Blue

		fmt.Println(style.Render("podlift") + " " + Version)
		fmt.Printf("Go version: %s\n", goVersion())
		fmt.Printf("OS/Arch: %s/%s\n", osName(), osArch())
		if Commit != "unknown" {
			fmt.Printf("Commit: %s\n", Commit)
		}
		if Date != "unknown" {
			fmt.Printf("Built: %s\n", Date)
		}
	},
}


// Helper functions (will be moved to proper package later)
func goVersion() string {
	// TODO: Get actual Go version at runtime
	return "go1.21+"
}

func osName() string {
	return runtime.GOOS
}

func osArch() string {
	return runtime.GOARCH
}

