package config

import (
	"slices"
	"testing"
)

func TestLoad_FallsBackToEmbeddedDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.UI.LogoEnabled {
		t.Errorf("ui.logo_enabled default should be false, got true")
	}
	if !slices.Contains(cfg.Worktree.Copy, ".env") {
		t.Errorf("worktree.copy missing .env sentinel; got %v", cfg.Worktree.Copy)
	}
	if len(cfg.Worktree.Symlink) == 0 {
		t.Errorf("worktree.symlink should not be empty in defaults")
	}
	if len(cfg.Worktree.Ignore) == 0 {
		t.Errorf("worktree.ignore should not be empty in defaults")
	}
	for _, want := range []string{"main", "master", "staging"} {
		if !slices.Contains(cfg.Branches.Protected, want) {
			t.Errorf("branches.protected default missing %q; got %v", want, cfg.Branches.Protected)
		}
	}
}

func TestIsProtected_ExactCaseSensitiveMatch(t *testing.T) {
	cfg := &Config{Branches: BranchesSection{Protected: []string{"main", "staging"}}}

	cases := []struct {
		branch string
		want   bool
	}{
		{"main", true},
		{"staging", true},
		{"Main", false},     // case differs
		{"main ", false},    // trailing space
		{"feat/main", false}, // substring
		{"develop", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := cfg.IsProtected(tc.branch); got != tc.want {
			t.Errorf("IsProtected(%q) = %v, want %v", tc.branch, got, tc.want)
		}
	}
}

func TestIsProtected_EmptyListNeverMatches(t *testing.T) {
	cfg := &Config{}
	if cfg.IsProtected("main") {
		t.Errorf("empty Protected list should not match anything")
	}
}

func TestSave_RoundTripsLogoToggle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load initial: %v", err)
	}
	cfg.UI.LogoEnabled = true
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if !reloaded.UI.LogoEnabled {
		t.Errorf("ui.logo_enabled did not persist; got false")
	}
	if !slices.Contains(reloaded.Worktree.Copy, ".env") {
		t.Errorf("worktree.copy lost on save round-trip; got %v", reloaded.Worktree.Copy)
	}
}
