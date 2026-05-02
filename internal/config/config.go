// Package config loads gitt's user configuration from ~/.gitt/config.toml.
// A missing file is not an error: callers transparently get the built-in
// defaults embedded at build time. The first time the user runs
// `gitt config`, EnsureFile materialises those defaults to disk so they
// can be edited.
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"

	"github.com/foreverfl/gitt/internal/paths"
)

//go:embed default.toml
var defaultTOML []byte

// Config mirrors config.toml. Sections are flat for now; add more as
// features land.
type Config struct {
	UI       UISection       `toml:"ui"`
	Worktree WorktreeSection `toml:"worktree"`
	Branches BranchesSection `toml:"branches"`
}

type UISection struct {
	LogoEnabled bool `toml:"logo_enabled"`
}

type WorktreeSection struct {
	Copy    []string `toml:"copy"`
	Symlink []string `toml:"symlink"`
	Ignore  []string `toml:"ignore"`
}

// BranchesSection holds branch-level policy. Protected lists branch names
// that gitt refuses to rename or remove — typically the long-lived trunk
// branches (main, master, staging) the user does not want to touch via
// `gitt`. Matching is case-sensitive and exact: git itself treats
// `Main` and `main` as different refs, so gitt mirrors that.
type BranchesSection struct {
	Protected []string `toml:"protected"`
}

// IsProtected reports whether branchName appears in the protected list.
// Exact, case-sensitive comparison — see the BranchesSection comment.
func (cfg *Config) IsProtected(branchName string) bool {
	return slices.Contains(cfg.Branches.Protected, branchName)
}

// Load returns the resolved config. If ~/.gitt/config.toml exists it is
// decoded and returned; otherwise the embedded defaults are returned.
// Callers should not assume which source they received — both shapes are
// fully populated.
func Load() (*Config, error) {
	path, err := paths.ConfigPath()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return decode(defaultTOML)
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	return decode(raw)
}

// Save writes cfg to ~/.gitt/config.toml, creating the directory if
// needed. Comments and formatting from any prior file are not
// preserved — gitt commands like `gitt logo` use this to flip a setting,
// and the canonical comment template lives in the embedded default.toml.
func Save(cfg *Config) error {
	path, err := paths.ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir config dir: %w", err)
	}
	raw, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

// EnsureFile creates ~/.gitt/config.toml from the embedded defaults if it
// does not yet exist, and returns the path either way. Used by
// `gitt config` so the editor always has something to open.
func EnsureFile() (string, error) {
	path, err := paths.ConfigPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat config %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("mkdir config dir: %w", err)
	}
	if err := os.WriteFile(path, defaultTOML, 0o644); err != nil {
		return "", fmt.Errorf("write config %s: %w", path, err)
	}
	return path, nil
}

func decode(raw []byte) (*Config, error) {
	var cfg Config
	if err := toml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return &cfg, nil
}
