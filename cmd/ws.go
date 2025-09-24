package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/okzmo/luna/internal/git"
	"github.com/okzmo/luna/internal/luna"
	"github.com/spf13/cobra"
)

var wsCmd = &cobra.Command{
	Use:   "ws <command>",
	Short: "Workspace operations",
	Long: `Manage Luna VCS workspaces for isolating your work.

Workspaces are git branches with metadata that track your progress.
All changes are automatically staged in workspaces.

For backward compatibility, you can also use:
  luna ws <name> <description>  # Same as luna ws create <name> <description>`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// For backward compatibility: luna ws <name> <description>
		if len(args) == 2 && args[0] != "create" && args[0] != "done" {
			name := args[0]
			description := args[1]

			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			gitFactory := git.NewRepositoryFactory()
			workspaceService := luna.NewWorkspaceService(gitFactory, wd)

			ctx := context.Background()
			if err := workspaceService.CreateWorkspace(ctx, wd, name, description); err != nil {
				return fmt.Errorf("failed to create workspace: %w", err)
			}

			fmt.Printf("Created workspace '%s' - %s\n", name, description)
			return nil
		}

		return cmd.Help()
	},
}

var wsCreateCmd = &cobra.Command{
	Use:   "create <name> <description>",
	Short: "Create a new workspace",
	Long: `Create a new workspace to isolate your work.

Examples:
  luna ws create feature-auth "Add user authentication"  
  luna ws create bugfix-login "Fix login validation issue"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description := args[1]

		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		gitFactory := git.NewRepositoryFactory()
		workspaceService := luna.NewWorkspaceService(gitFactory, wd)

		ctx := context.Background()
		if err := workspaceService.CreateWorkspace(ctx, wd, name, description); err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		fmt.Printf("Created workspace '%s' - %s\n", name, description)
		return nil
	},
}

var wsDoneCmd = &cobra.Command{
	Use:   "done",
	Short: "Finish current workspace",
	Long: `Finish the current workspace by squashing all commits and rebasing onto the luna branch.

This will:
- Squash all commits in the workspace into one commit
- Use the workspace description as the commit message
- Rebase the squashed commit onto the luna branch
- Delete the workspace branch
- Switch back to the luna branch

Example:
  luna ws done`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		gitFactory := git.NewRepositoryFactory()
		workspaceService := luna.NewWorkspaceService(gitFactory, wd)

		ctx := context.Background()
		if err := workspaceService.FinishWorkspace(ctx, wd); err != nil {
			return fmt.Errorf("failed to finish workspace: %w", err)
		}

		fmt.Println("Workspace completed and merged to luna branch")
		return nil
	},
}

func init() {
	wsCmd.AddCommand(wsCreateCmd)
	wsCmd.AddCommand(wsDoneCmd)
	rootCmd.AddCommand(wsCmd)
}
