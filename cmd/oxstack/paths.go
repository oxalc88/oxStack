package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// Standard paths
func claudeSkillsDir() string  { return filepath.Join(homeDir(), ".claude", "skills") }
func agentsSkillsDir() string  { return filepath.Join(homeDir(), ".agents", "skills") }
func claudeAgentsDir() string  { return filepath.Join(homeDir(), ".claude", "agents") }
func claudeSettingsPath() string { return filepath.Join(homeDir(), ".claude", "settings.json") }
func codexDir() string         { return filepath.Join(homeDir(), ".codex") }
func codexConfigPath() string  { return filepath.Join(homeDir(), ".codex", "config.toml") }
func opencodeDir() string      { return filepath.Join(homeDir(), ".opencode") }
func gstackDir() string        { return filepath.Join(homeDir(), ".claude", "skills", "gstack") }

func syncFilePath() string {
	return filepath.Join(repoRoot(), "gstack-ab", ".last-sync-commit")
}

// symlink creates a symlink at dst pointing to src.
// On Windows, falls back to copying if symlink fails (needs admin).
func symlink(src, dst string) error {
	// Remove existing symlink or backup existing real file/dir
	if info, err := os.Lstat(dst); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			os.Remove(dst)
		} else {
			bak := dst + ".bak"
			warnf("Backing up %s → %s", dst, bak)
			if err := os.Rename(dst, bak); err != nil {
				return fmt.Errorf("backup %s: %w", dst, err)
			}
		}
	}

	err := os.Symlink(src, dst)
	if err != nil && runtime.GOOS == "windows" {
		// Fallback: copy on Windows
		warnf("Symlink failed, copying instead: %s → %s", src, dst)
		return copyDir(src, dst)
	}
	return err
}

// copyDir recursively copies src to dst.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			s := filepath.Join(src, entry.Name())
			d := filepath.Join(dst, entry.Name())
			if err := copyDir(s, d); err != nil {
				return err
			}
		}
		return nil
	}

	return copyFile(src, dst)
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, srcInfo.Mode())
}
