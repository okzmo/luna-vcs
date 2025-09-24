package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/okzmo/luna/internal/git"
	"github.com/okzmo/luna/internal/luna"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <description>",
	Short: "Create a new step in the current workspace",
	Long: `Create a new step in your current workspace.

This commits your current work and starts a new step.
All changes are automatically staged before committing.

Examples:
  luna new "Add login form validation"
  luna new "Fix CSS styling issues"
  luna new "Implement user registration"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := args[0]

		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		gitFactory := git.NewRepositoryFactory()
		workspaceService := luna.NewWorkspaceService(gitFactory, wd)

		ctx := context.Background()
		if err := workspaceService.CreateStep(ctx, wd, description); err != nil {
			return fmt.Errorf("failed to create step: %w", err)
		}

		fmt.Printf("Created step: %s\n", description)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
