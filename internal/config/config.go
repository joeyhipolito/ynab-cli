// Package config handles reading and writing the YNAB CLI configuration file.
// Configuration is stored in ~/.ynab/config in INI-style format.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ConfigDir is the directory name for YNAB configuration.
	ConfigDir = ".ynab"
	// ConfigFile is the configuration file name.
	ConfigFile = "config"
)

// Config represents the YNAB CLI configuration.
type Config struct {
	AccessToken     string
	DefaultBudgetID string
	APIBaseURL      string
}

// Path returns the full path to the config file (~/.ynab/config).
func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ConfigDir, ConfigFile)
}

// Dir returns the full path to the config directory (~/.ynab/).
func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ConfigDir)
}

// Load reads the configuration from ~/.ynab/config.
// Returns an empty Config (not an error) if the file doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{}
	path := Path()
	if path == "" {
		return cfg, nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "access_token":
			cfg.AccessToken = value
		case "default_budget_id":
			cfg.DefaultBudgetID = value
		case "api_base_url":
			cfg.APIBaseURL = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to ~/.ynab/config with proper permissions.
func Save(cfg *Config) error {
	dir := Dir()
	if dir == "" {
		return fmt.Errorf("cannot determine home directory")
	}

	// Create config directory with 700 permissions
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path := Path()

	// Build config content
	var b strings.Builder
	b.WriteString("# YNAB CLI Configuration\n")
	b.WriteString("# Created by: ynab-cli configure\n")
	b.WriteString("\n")
	b.WriteString("# Your YNAB Personal Access Token\n")
	b.WriteString("# Get from: https://app.ynab.com/settings/developer\n")
	fmt.Fprintf(&b, "access_token=%s\n", cfg.AccessToken)
	b.WriteString("\n")
	b.WriteString("# Default budget ID\n")
	fmt.Fprintf(&b, "default_budget_id=%s\n", cfg.DefaultBudgetID)
	b.WriteString("\n")
	b.WriteString("# API base URL\n")
	if cfg.APIBaseURL != "" {
		fmt.Fprintf(&b, "api_base_url=%s\n", cfg.APIBaseURL)
	} else {
		b.WriteString("api_base_url=https://api.youneedabudget.com/v1\n")
	}

	// Write file with 600 permissions
	if err := os.WriteFile(path, []byte(b.String()), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists returns true if the config file exists.
func Exists() bool {
	path := Path()
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// Permissions returns the file permissions of the config file, or an error.
func Permissions() (os.FileMode, error) {
	path := Path()
	if path == "" {
		return 0, fmt.Errorf("cannot determine config path")
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
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
