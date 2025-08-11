package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information - will be set during build
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print version information for AI Dependency Manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AI Dependency Manager (AutoUpdateAgent)\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
