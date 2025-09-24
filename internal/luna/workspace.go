package luna

import (
	"context"
	"fmt"
	"time"

	"github.com/okzmo/luna/internal/git"
)

type WorkspaceService struct {
	gitFactory      git.RepositoryFactory
	metadataService *MetadataService
}

func NewWorkspaceService(gitFactory git.RepositoryFactory, repoPath string) *WorkspaceService {
	return &WorkspaceService{
		gitFactory:      gitFactory,
		metadataService: NewMetadataService(repoPath),
	}
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, repoPath, name, description string) error {
	repo := s.gitFactory.NewRepository(repoPath)

	isRepo, err := repo.IsRepository(repoPath)
	if err != nil {
		return fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return fmt.Errorf("not a luna repository")
	}

	if err := repo.CreateBranch(ctx, name, "luna"); err != nil {
		return fmt.Errorf("failed to create workspace branch: %w", err)
	}

	if err := repo.SwitchBranch(ctx, name); err != nil {
		return fmt.Errorf("failed to switch to workspace: %w", err)
	}

	if err := s.metadataService.CreateWorkspace(name, description); err != nil {
		return fmt.Errorf("failed to create workspace metadata: %w", err)
	}

	return nil
}

func (s *WorkspaceService) CreateStep(ctx context.Context, repoPath, description string) error {
	repo := s.gitFactory.NewRepository(repoPath)
	metadata, err := s.metadataService.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	currentWorkspace := metadata.CurrentWorkspace
	if currentWorkspace == "" {
		return fmt.Errorf("no active workspace - create one with 'luna ws <name> <description>'")
	}

	if err := repo.StageAll(ctx); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	nbOfSteps := len(metadata.Workspaces[currentWorkspace].Steps)

	var lastStepDescription string
	if nbOfSteps > 0 {
		lastStepDescription = metadata.Workspaces[currentWorkspace].Steps[nbOfSteps-1].Description
	} else {
		lastStepDescription = metadata.Workspaces[currentWorkspace].Description
	}

	commitHash, err := repo.Commit(ctx, lastStepDescription)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	if err := s.metadataService.AddStep(description, commitHash); err != nil {
		return fmt.Errorf("failed to add step metadata: %w", err)
	}

	return nil
}

func (s *WorkspaceService) FinishWorkspace(ctx context.Context, repoPath string) error {
	currentWorkspace, err := s.metadataService.GetCurrentWorkspace()
	if err != nil {
		return fmt.Errorf("failed to get current workspace: %w", err)
	}

	if currentWorkspace == "" {
		return fmt.Errorf("no active workspace")
	}

	metadata, err := s.metadataService.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	workspace, exists := metadata.Workspaces[currentWorkspace]
	if !exists {
		return fmt.Errorf("workspace '%s' not found", currentWorkspace)
	}

	repo := s.gitFactory.NewRepository(repoPath)

	// Stage and commit any pending changes before squashing
	if err := repo.StageAll(ctx); err != nil {
		return fmt.Errorf("failed to stage pending changes: %w", err)
	}

	// Check if there are changes to commit
	hasChanges, err := repo.HasStagedChanges(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for staged changes: %w", err)
	}

	if hasChanges {
		// Get the last step description for the final commit message
		var finalStepDescription string
		if len(workspace.Steps) > 0 {
			finalStepDescription = workspace.Steps[len(workspace.Steps)-1].Description
		} else {
			finalStepDescription = "Final changes"
		}

		commitHash, err := repo.Commit(ctx, finalStepDescription)
		if err != nil {
			return fmt.Errorf("failed to commit pending changes: %w", err)
		}

		// Add this final commit to the metadata
		finalStep := Step{
			Description: "Final workspace changes",
			CommitHash:  commitHash,
			CreatedAt:   time.Now(),
		}
		workspace.Steps = append(workspace.Steps, finalStep)
		metadata.Workspaces[currentWorkspace] = workspace
	}

	if err := repo.SquashAndRebase(ctx, "luna", workspace.Description); err != nil {
		return fmt.Errorf("failed to squash and rebase workspace: %w", err)
	}

	delete(metadata.Workspaces, currentWorkspace)
	metadata.CurrentWorkspace = ""

	if err := s.metadataService.SaveMetadata(metadata); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}
