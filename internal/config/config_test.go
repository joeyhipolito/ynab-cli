package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joeyhipolito/ynab-cli/internal/config"
)

// TestPathDefaultsToHome verifies that Path() falls back to ~/.ynab/config when
// YNAB_CONFIG_DIR is not set.
func TestPathDefaultsToHome(t *testing.T) {
	t.Setenv("YNAB_CONFIG_DIR", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	got := config.Path()
	want := filepath.Join(home, ".ynab", "config")
	if got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

// TestPathUsesEnvVar verifies that YNAB_CONFIG_DIR overrides the default directory.
func TestPathUsesEnvVar(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("YNAB_CONFIG_DIR", tmp)

	got := config.Path()
	want := filepath.Join(tmp, "config")
	if got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

// TestDirUsesEnvVar verifies that Dir() respects YNAB_CONFIG_DIR.
func TestDirUsesEnvVar(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("YNAB_CONFIG_DIR", tmp)

	got := config.Dir()
	if got != tmp {
		t.Errorf("Dir() = %q, want %q", got, tmp)
	}
}

// TestLoadSaveRoundtrip writes a config via Save and reads it back via Load.
func TestLoadSaveRoundtrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("YNAB_CONFIG_DIR", tmp)

	cfg := &config.Config{
		AccessToken:     "tok-abc",
		DefaultBudgetID: "budget-123",
		APIBaseURL:      "https://api.youneedabudget.com/v1",
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.AccessToken != cfg.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, cfg.AccessToken)
	}
	if got.DefaultBudgetID != cfg.DefaultBudgetID {
		t.Errorf("DefaultBudgetID = %q, want %q", got.DefaultBudgetID, cfg.DefaultBudgetID)
	}
	if got.APIBaseURL != cfg.APIBaseURL {
		t.Errorf("APIBaseURL = %q, want %q", got.APIBaseURL, cfg.APIBaseURL)
	}
}

// TestLoadMissingFileReturnsEmpty verifies Load returns empty Config when no file exists.
func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("YNAB_CONFIG_DIR", tmp)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AccessToken != "" || cfg.DefaultBudgetID != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}
