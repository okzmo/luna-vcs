package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "luna",
	Short: "A brief description of your application",
	Long: `Luna VCS - A safer, more intuitive version control system built on git.

Luna VCS provides enhanced safety features and user-friendly workflows
while using git as the underlying data layer.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}
