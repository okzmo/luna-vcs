package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type gitRepository struct {
	path string
	repo *git.Repository
}

type gitRepositoryFactory struct{}

func NewRepositoryFactory() RepositoryFactory {
	return &gitRepositoryFactory{}
}

func (f *gitRepositoryFactory) NewRepository(path string) Repository {
	return &gitRepository{
		path: path,
	}
}

func (r *gitRepository) Init(ctx context.Context, path string) error {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	r.path = absPath

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", absPath, err)
		}
	}

	isRepo, err := r.IsRepository(absPath)
	if err != nil {
		return fmt.Errorf("failed to check if directory is already a repository: %w", err)
	}
	if isRepo {
		return fmt.Errorf("directory %s is already a luna repository", absPath)
	}

	repo, err := git.PlainInit(absPath, false)
	if err != nil {
		return fmt.Errorf("failed to initialize luna repository at %s: %w", absPath, err)
	}

	r.repo = repo

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	gitignorePath := filepath.Join(absPath, ".gitignore")
	gitignoreContent := `# Luna files
.luna/

# Common files to ignore
*.log
*.tmp
*~
.DS_Store
Thumbs.db
`
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore file: %w", err)
	}

	_, err = worktree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add files to staging: %w", err)
	}

	signature := &object.Signature{
		Name:  "Luna",
		Email: "luna@vcs.local",
	}

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: signature,
	})
	if err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("luna"),
		Create: true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout luna branch: %w", err)
	}

	return nil
}

func (r *gitRepository) IsRepository(path string) (bool, error) {
	gitDir := filepath.Join(path, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	_, err := git.PlainOpen(path)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (r *gitRepository) GetPath() string {
	return r.path
}

func (r *gitRepository) CreateBranch(ctx context.Context, branchName, baseBranch string) error {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	baseBranchRef, err := repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
	if err != nil {
		return fmt.Errorf("failed to get base branch reference: %w", err)
	}

	newBranchRef := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(newBranchRef, baseBranchRef.Hash())

	if err := repo.Storer.SetReference(ref); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

func (r *gitRepository) SwitchBranch(ctx context.Context, branchName string) error {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

func (r *gitRepository) StageAll(ctx context.Context) error {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = worktree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	return nil
}

func (r *gitRepository) Commit(ctx context.Context, message string) (string, error) {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	signature, err := r.GetUserSignature()
	if err != nil {
		return "", fmt.Errorf("failed to get user signature: %w", err)
	}

	commit, err := worktree.Commit(message, &git.CommitOptions{
		Author: signature,
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return commit.String(), nil
}

func (r *gitRepository) GetCurrentBranch(ctx context.Context) (string, error) {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	if !head.Name().IsBranch() {
		return "", fmt.Errorf("HEAD is not pointing to a branch")
	}

	return head.Name().Short(), nil
}

func (r *gitRepository) SquashAndRebase(ctx context.Context, baseBranch, commitMessage string) error {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	currentBranch, err := r.GetCurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch == baseBranch {
		return fmt.Errorf("already on base branch %s", baseBranch)
	}

	baseBranchRef, err := repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
	if err != nil {
		return fmt.Errorf("failed to get base branch reference: %w", err)
	}

	workspaceBranchRef, err := repo.Reference(plumbing.NewBranchReferenceName(currentBranch), true)
	if err != nil {
		return fmt.Errorf("failed to get workspace branch reference: %w", err)
	}

	workspaceCommitObj, err := repo.CommitObject(workspaceBranchRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get workspace commit: %w", err)
	}

	signature, err := r.GetUserSignature()
	if err != nil {
		return fmt.Errorf("failed to get user signature: %w", err)
	}

	squashedCommit := &object.Commit{
		Author:       *signature,
		Committer:    *signature,
		Message:      commitMessage,
		TreeHash:     workspaceCommitObj.TreeHash,
		ParentHashes: []plumbing.Hash{baseBranchRef.Hash()},
	}

	obj := repo.Storer.NewEncodedObject()
	if err := squashedCommit.Encode(obj); err != nil {
		return fmt.Errorf("failed to encode squashed commit: %w", err)
	}

	commitHash, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("failed to store squashed commit: %w", err)
	}

	newRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(baseBranch), commitHash)
	if err := repo.Storer.SetReference(newRef); err != nil {
		return fmt.Errorf("failed to update base branch reference: %w", err)
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(baseBranch),
	}); err != nil {
		return fmt.Errorf("failed to checkout base branch: %w", err)
	}

	if err := repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(currentBranch)); err != nil {
		return fmt.Errorf("failed to delete workspace branch: %w", err)
	}

	return nil
}

func (r *gitRepository) HasStagedChanges(ctx context.Context) (bool, error) {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree status: %w", err)
	}

	// Check if there are any staged changes (index modifications)
	for _, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified {
			return true, nil
		}
	}

	return false, nil
}

func (r *gitRepository) GetUserSignature() (*object.Signature, error) {
	// Get user name from git config
	nameCmd := exec.Command("git", "config", "--global", "user.name")
	nameOutput, err := nameCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get user.name from git config: %w", err)
	}
	name := strings.TrimSpace(string(nameOutput))
	if name == "" {
		return nil, fmt.Errorf("user.name is not set in git config")
	}

	// Get user email from git config
	emailCmd := exec.Command("git", "config", "--global", "user.email")
	emailOutput, err := emailCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get user.email from git config: %w", err)
	}
	email := strings.TrimSpace(string(emailOutput))
	if email == "" {
		return nil, fmt.Errorf("user.email is not set in git config")
	}

	return &object.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}, nil
}
