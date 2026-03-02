// Package config handles reading and writing the YNAB CLI configuration file.
// The config directory defaults to ~/.ynab but can be overridden via YNAB_CONFIG_DIR.
package config

import (
	"fmt"
	"os"

	sharedconfig "github.com/joeyhipolito/publishing-shared/config"
)

const (
	// ConfigDir is the directory name for YNAB configuration.
	ConfigDir = ".ynab"
	// ConfigFile is the configuration file name.
	ConfigFile = "config"
)

var store = sharedconfig.NewStoreWithEnv(ConfigDir, "YNAB_CONFIG_DIR")

// Config represents the YNAB CLI configuration.
type Config struct {
	AccessToken     string
	DefaultBudgetID string
	APIBaseURL      string
}

// Path returns the full path to the config file (~/.ynab/config or $YNAB_CONFIG_DIR/config).
// Returns "" if the path cannot be determined.
func Path() string {
	p, err := store.Path()
	if err != nil {
		return ""
	}
	return p
}

// Dir returns the full path to the config directory.
// Returns "" if the path cannot be determined.
func Dir() string {
	d, err := store.BaseDir()
	if err != nil {
		return ""
	}
	return d
}

// Load reads the configuration from the config file.
// Returns an empty Config (not an error) if the file doesn't exist.
func Load() (*Config, error) {
	values, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &Config{
		AccessToken:     values["access_token"],
		DefaultBudgetID: values["default_budget_id"],
		APIBaseURL:      values["api_base_url"],
	}, nil
}

// Save writes the configuration to the config file with proper permissions.
func Save(cfg *Config) error {
	values := map[string]string{
		"access_token":     cfg.AccessToken,
		"default_budget_id": cfg.DefaultBudgetID,
	}
	if cfg.APIBaseURL != "" {
		values["api_base_url"] = cfg.APIBaseURL
	} else {
		values["api_base_url"] = "https://api.youneedabudget.com/v1"
	}

	header := "# YNAB CLI Configuration\n# Created by: ynab-cli configure\n\n# Your YNAB Personal Access Token\n# Get from: https://app.ynab.com/settings/developer"
	keyOrder := []string{"access_token", "default_budget_id", "api_base_url"}
	return store.Save(values, header, keyOrder)
}

// Exists returns true if the config file exists.
func Exists() bool {
	return store.Exists()
}

// Permissions returns the file permissions of the config file, or an error.
func Permissions() (os.FileMode, error) {
	return store.Permissions()
}

// ResolveToken returns the access token using config priority:
// config file > environment variable.
func ResolveToken() string {
	cfg, err := Load()
	if err == nil && cfg.AccessToken != "" {
		return cfg.AccessToken
	}
	return os.Getenv("YNAB_ACCESS_TOKEN")
}

// ResolveBudgetID returns the default budget ID from config or environment.
func ResolveBudgetID() string {
	cfg, err := Load()
	if err == nil && cfg.DefaultBudgetID != "" {
		return cfg.DefaultBudgetID
	}
	return os.Getenv("YNAB_DEFAULT_BUDGET_ID")
}
