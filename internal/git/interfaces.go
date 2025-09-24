// Package git provides interfaces and implementations for git operations in Luna.
// It's for easy testing and mocking.
package git

import (
	"context"

	"github.com/go-git/go-git/v6/plumbing/object"
)

// Repository represents a git repository interface for testability.
type Repository interface {
	// Init initializes a new git repository at the specified path.
	// Returns an error if initialization fails or if a repository already exists.
	Init(ctx context.Context, path string) error

	// IsRepository checks if the given path is already a git repository.
	IsRepository(path string) (bool, error)

	// GetPath returns the root path of the repository.
	GetPath() string

	// CreateBranch creates a new branch from the specified base branch.
	CreateBranch(ctx context.Context, branchName, baseBranch string) error

	// SwitchBranch switches to the specified branch.
	SwitchBranch(ctx context.Context, branchName string) error

	// StageAll stages all changes in the working directory.
	StageAll(ctx context.Context) error

	// Commit creates a commit with the given message and returns the commit hash.
	Commit(ctx context.Context, message string) (string, error)

	// GetCurrentBranch returns the name of the current branch.
	GetCurrentBranch(ctx context.Context) (string, error)

	// SquashAndRebase squashes all commits from current branch since baseBranch and rebases onto baseBranch.
	SquashAndRebase(ctx context.Context, baseBranch, commitMessage string) error

	// HasStagedChanges checks if there are any staged changes ready to commit.
	HasStagedChanges(ctx context.Context) (bool, error)

	// GetUserSignature returns the user's git signature from global config.
	GetUserSignature() (*object.Signature, error)
}

// RepositoryFactory creates Repository instances.
type RepositoryFactory interface {
	// NewRepository creates a new Repository instance for the given path.
	NewRepository(path string) Repository
}
