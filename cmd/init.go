package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/okzmo/luna/internal/git"
	"github.com/okzmo/luna/internal/luna"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new Luna VCS repository",
	Long: `Initialize a new Luna VCS repository in the specified directory.
If no path is provided, initializes in the current directory.

Luna VCS creates a git repository as the underlying data layer and adds
its own structures on top to provide enhanced safety and workflow features.

Examples:
  luna init                    # Initialize in current directory
  luna init my-project         # Initialize in ./my-project
  luna init /path/to/project   # Initialize in absolute path`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string
		if len(args) > 0 {
			path = args[0]
		}

		gitFactory := git.NewRepositoryFactory()
		initService := luna.NewInitService(gitFactory)

		ctx := context.Background()
		if err := initService.InitRepository(ctx, path); err != nil {
			return fmt.Errorf("failed to initialize Luna repository: %w", err)
		}

		repo := gitFactory.NewRepository(path)
		actualPath := repo.GetPath()
		if actualPath == "" {
			if wd, err := os.Getwd(); err == nil {
				actualPath = wd
			} else {
				actualPath = "current directory"
			}
		}

		fmt.Printf("Initialized Luna repository in %s\n", actualPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
