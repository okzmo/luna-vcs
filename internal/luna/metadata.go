package luna

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type WorkspaceMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Steps       []Step    `json:"steps"`
}

type Step struct {
	Description string    `json:"description"`
	CommitHash  string    `json:"commit_hash"`
	CreatedAt   time.Time `json:"created_at"`
}

type LunaMetadata struct {
	Workspaces       map[string]WorkspaceMetadata `json:"workspaces"`
	CurrentWorkspace string                       `json:"current_workspace"`
}

type MetadataService struct {
	repoPath string
}

func NewMetadataService(repoPath string) *MetadataService {
	return &MetadataService{
		repoPath: repoPath,
	}
}

func (m *MetadataService) getMetadataPath() string {
	return filepath.Join(m.repoPath, ".git", "metadata.json")
}

func (m *MetadataService) LoadMetadata() (*LunaMetadata, error) {
	metadataPath := m.getMetadataPath()

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return &LunaMetadata{
			Workspaces:       make(map[string]WorkspaceMetadata),
			CurrentWorkspace: "",
		}, nil
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata LunaMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	if metadata.Workspaces == nil {
		metadata.Workspaces = make(map[string]WorkspaceMetadata)
	}

	return &metadata, nil
}

func (m *MetadataService) SaveMetadata(metadata *LunaMetadata) error {
	metadataPath := m.getMetadataPath()

	gitDir := filepath.Dir(metadataPath)
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return fmt.Errorf("failed to create .git directory: %w", err)
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

func (m *MetadataService) CreateWorkspace(name, description string) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if _, exists := metadata.Workspaces[name]; exists {
		return fmt.Errorf("workspace '%s' already exists", name)
	}

	workspace := WorkspaceMetadata{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		Steps:       []Step{},
	}

	metadata.Workspaces[name] = workspace
	metadata.CurrentWorkspace = name

	return m.SaveMetadata(metadata)
}

func (m *MetadataService) AddStep(description, commitHash string) error {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if metadata.CurrentWorkspace == "" {
		return fmt.Errorf("no active workspace")
	}

	workspace, exists := metadata.Workspaces[metadata.CurrentWorkspace]
	if !exists {
		return fmt.Errorf("current workspace '%s' not found", metadata.CurrentWorkspace)
	}

	step := Step{
		Description: description,
		CommitHash:  commitHash,
		CreatedAt:   time.Now(),
	}

	workspace.Steps = append(workspace.Steps, step)
	metadata.Workspaces[metadata.CurrentWorkspace] = workspace

	return m.SaveMetadata(metadata)
}

func (m *MetadataService) GetCurrentWorkspace() (string, error) {
	metadata, err := m.LoadMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to load metadata: %w", err)
	}

	return metadata.CurrentWorkspace, nil
}
