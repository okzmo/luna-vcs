// Package luna provides the core Luna VCS services and business logic.
package luna

import (
	"context"
	"fmt"

	"github.com/okzmo/luna/internal/git"
)

type InitService struct {
	gitFactory git.RepositoryFactory
}

func NewInitService(gitFactory git.RepositoryFactory) *InitService {
	return &InitService{
		gitFactory: gitFactory,
	}
}

func (s *InitService) InitRepository(ctx context.Context, path string) error {
	repo := s.gitFactory.NewRepository(path)

	if err := repo.Init(ctx, path); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	return nil
}
