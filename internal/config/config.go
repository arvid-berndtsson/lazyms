package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents user configuration.
type Config struct {
	TenantID      string `yaml:"tenantId"`
	ClientID      string `yaml:"clientId"`
	PreferredAuth string `yaml:"preferredAuth"`
}

// Load reads configuration from ~/.config/lazyms/config.yaml.
// Missing file is not an error; defaults are returned.
func Load() (Config, error) {
	var cfg Config
	path := filepath.Join(userConfigDir(), "lazyms", "config.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// defaults
			if cfg.PreferredAuth == "" {
				cfg.PreferredAuth = "cli"
			}
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	if cfg.PreferredAuth == "" {
		cfg.PreferredAuth = "cli"
	}
	return cfg, nil
}

func userConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(h, ".config")
}
