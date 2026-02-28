package gitai

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// RepoInfo holds git repository information
type RepoInfo struct {
	Path      string
	IsRepo    bool
	Head      string
	Branch    string
	Commit    string
	Worktree  string
}

// GetGitRepoInfo retrieves current git repository information
func GetGitRepoInfo() (*RepoInfo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Find git root
	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return &RepoInfo{
				Path:   cwd,
				IsRepo: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to open git repo: %w", err)
	}

	// Get worktree path
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get branch name
	branch := head.Name().Short()
	if head.Name() == plumbing.HEAD {
		// Detached HEAD
		branch = "HEAD"
	}

	info := &RepoInfo{
		Path:     worktree.Filesystem.Root(),
		IsRepo:   true,
		Head:     head.Name().String(),
		Branch:   branch,
		Commit:   head.Hash().String(),
		Worktree: cwd,
	}

	return info, nil
}

// MustGetGitRepo retrieves git repo info, exits if not in a git repo
func MustGetGitRepo() (*git.Repository, string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repository (or any parent up to mount point)\n")
		os.Exit(1)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return repo, worktree.Filesystem.Root()
}

// GetGitRoot returns the git root directory
func GetGitRoot() (string, error) {
	repo, root, err := getRepoAndRoot()
	if err != nil {
		return "", err
	}
	_ = repo
	return root, nil
}

func getRepoAndRoot() (*git.Repository, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, "", err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, "", err
	}

	return repo, worktree.Filesystem.Root(), nil
}

// GetRelativePath returns path relative to git root
func GetRelativePath(absPath string) (string, error) {
	root, err := GetGitRoot()
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(root, absPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// GetAbsolutePath returns absolute path from git root relative path
func GetAbsolutePath(relPath string) (string, error) {
	root, err := GetGitRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, relPath), nil
}
