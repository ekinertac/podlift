package commands

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags
	// If not set, it will be detected from git
	Version = ""
	Commit  = ""
	Date    = ""
)

var rootCmd = &cobra.Command{
	Use:   "podlift",
	Short: "Simple, transparent deployment for containerized applications",
	Long: `podlift deploys your Docker containers to any server with SSH access.
No black boxes, no magic, no broken promises.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       getVersion(),
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add version subcommand for 'podlift version'
	rootCmd.AddCommand(versionCmd)
	
	// Customize --version flag output
	rootCmd.SetVersionTemplate(getVersionOutput())
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show podlift version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(getVersionOutput())
	},
}

// getVersion returns the version string, detecting from git if not set at build time
func getVersion() string {
	if Version != "" {
		return Version
	}
	
	// Try to get version from git tag
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	
	return "dev"
}

// getVersionOutput returns formatted version information
func getVersionOutput() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")) // Blue

	version := getVersion()
	output := style.Render("podlift") + " " + version + "\n"
	output += fmt.Sprintf("Go version: %s\n", runtime.Version())
	output += fmt.Sprintf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	
	if Commit != "" && Commit != "unknown" {
		output += fmt.Sprintf("Commit: %s\n", Commit)
	}
	if Date != "" && Date != "unknown" {
		output += fmt.Sprintf("Built: %s\n", Date)
	}
	
	return output
}

