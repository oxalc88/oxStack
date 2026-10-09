package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// homeDir returns the user's home directory.
func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "x Could not determine home directory: %v\n", err)
		os.Exit(1)
	}
	return h
}

// repoRoot returns the oxStack repository root.
// It resolves relative to the executable, following symlinks.
func repoRoot() string {
	exe, err := os.Executable()
	if err != nil {
		// Fallback: check if we're running from the repo
		if wd, e := os.Getwd(); e == nil {
			if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
				return wd
			}
		}
		fmt.Fprintf(os.Stderr, "x Could not determine executable path: %v\n", err)
		os.Exit(1)
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "x Could not resolve executable symlinks: %v\n", err)
		os.Exit(1)
	}

	// The binary is at cmd/oxstack/ relative to repo root, or at repo root,
	// or installed globally. Try to find go.mod by walking up.
	dir := filepath.Dir(exe)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// If installed globally (go install), look for OXSTACK_ROOT env var
	if root := os.Getenv("OXSTACK_ROOT"); root != "" {
		return root
	}

	// Final fallback: current working directory
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
	}

	fmt.Fprintf(os.Stderr, "x Could not find oxStack repo root. Set OXSTACK_ROOT or run from the repo.\n")
	os.Exit(1)
	return ""
}

// Maintenance sources live outside every runtime's skill discovery directories.
func gstackDir() string {
	if source := os.Getenv("OXSTACK_GSTACK_SOURCE"); source != "" {
		return source
	}
	return filepath.Join(repoRoot(), "catalog", "gstack")
}

func syncFilePath() string {
	return filepath.Join(repoRoot(), "gstack-ab", ".last-sync-commit")
}

// copyFile is used only for catalog maintenance, never implicit activation.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode().Perm())
}
