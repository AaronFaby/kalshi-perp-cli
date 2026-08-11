package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	EnvDemo = "demo"
	EnvProd = "prod"
)

// Config holds runtime configuration for the CLI.
type Config struct {
	Env            string `yaml:"env"`
	APIKey         string `yaml:"api_key"`
	PrivateKeyPath string `yaml:"private_key_path"`
	PrivateKeyPEM  string `yaml:"-"` // from KALSHI_PRIVATE_KEY env only
	BaseURL        string `yaml:"base_url"`
	WSURL          string `yaml:"ws_url"`
	Format         string `yaml:"format"`
	TimeoutSec     int    `yaml:"timeout_sec"`
	ConfigPath     string `yaml:"-"`
	Verbose        bool   `yaml:"-"`
}

// File is the on-disk YAML shape.
type File struct {
	Env            string `yaml:"env"`
	APIKey         string `yaml:"api_key"`
	PrivateKeyPath string `yaml:"private_key_path"`
	BaseURL        string `yaml:"base_url"`
	WSURL          string `yaml:"ws_url"`
	Format         string `yaml:"format"`
	TimeoutSec     int    `yaml:"timeout_sec"`
}

// DefaultConfigPath returns ~/.config/kalshi-perp/config.yaml (XDG-aware).
func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "kalshi-perp", "config.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(home, ".config", "kalshi-perp", "config.yaml")
}

// Load merges file, environment, and flag overrides.
// Flag values that are non-empty (or non-zero where applicable) win.
func Load(cfgPath string, flags Config) (Config, error) {
	if cfgPath == "" {
		cfgPath = DefaultConfigPath()
	}

	cfg := Config{
		Env:        EnvDemo,
		Format:     "table",
		TimeoutSec: 30,
		ConfigPath: cfgPath,
	}

	if data, err := os.ReadFile(cfgPath); err == nil {
		var f File
		if err := yaml.Unmarshal(data, &f); err != nil {
			return cfg, fmt.Errorf("parse config %s: %w", cfgPath, err)
		}
		mergeFile(&cfg, f)
	} else if !os.IsNotExist(err) {
		return cfg, fmt.Errorf("read config %s: %w", cfgPath, err)
	}

	mergeEnv(&cfg)
	mergeFlags(&cfg, flags)

	cfg.Env = strings.ToLower(strings.TrimSpace(cfg.Env))
	if cfg.Env != EnvDemo && cfg.Env != EnvProd {
		return cfg, fmt.Errorf("invalid env %q (want demo or prod)", cfg.Env)
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL(cfg.Env)
	}
	if cfg.WSURL == "" {
		cfg.WSURL = DefaultWSURL(cfg.Env)
	}
	if cfg.Format == "" {
		cfg.Format = "table"
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 30
	}

	return cfg, nil
}

func mergeFile(cfg *Config, f File) {
	if f.Env != "" {
		cfg.Env = f.Env
	}
	if f.APIKey != "" {
		cfg.APIKey = f.APIKey
	}
	if f.PrivateKeyPath != "" {
		cfg.PrivateKeyPath = f.PrivateKeyPath
	}
	if f.BaseURL != "" {
		cfg.BaseURL = f.BaseURL
	}
	if f.WSURL != "" {
		cfg.WSURL = f.WSURL
	}
	if f.Format != "" {
		cfg.Format = f.Format
	}
	if f.TimeoutSec > 0 {
		cfg.TimeoutSec = f.TimeoutSec
	}
}

func mergeEnv(cfg *Config) {
	if v := os.Getenv("KALSHI_ENV"); v != "" {
		cfg.Env = v
	}
	if v := os.Getenv("KALSHI_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("KALSHI_PRIVATE_KEY_PATH"); v != "" {
		cfg.PrivateKeyPath = v
	}
	if v := os.Getenv("KALSHI_PRIVATE_KEY"); v != "" {
		cfg.PrivateKeyPEM = v
	}
	if v := os.Getenv("KALSHI_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("KALSHI_WS_URL"); v != "" {
		cfg.WSURL = v
	}
}

func mergeFlags(cfg *Config, flags Config) {
	if flags.Env != "" {
		cfg.Env = flags.Env
	}
	if flags.APIKey != "" {
		cfg.APIKey = flags.APIKey
	}
	if flags.PrivateKeyPath != "" {
		cfg.PrivateKeyPath = flags.PrivateKeyPath
	}
	if flags.BaseURL != "" {
		cfg.BaseURL = flags.BaseURL
	}
	if flags.WSURL != "" {
		cfg.WSURL = flags.WSURL
	}
	if flags.Format != "" {
		cfg.Format = flags.Format
	}
	if flags.TimeoutSec > 0 {
		cfg.TimeoutSec = flags.TimeoutSec
	}
	if flags.Verbose {
		cfg.Verbose = true
	}
	if flags.ConfigPath != "" {
		cfg.ConfigPath = flags.ConfigPath
	}
}

// DefaultBaseURL returns the REST base URL for env.
func DefaultBaseURL(env string) string {
	if env == EnvProd {
		return "https://external-api.kalshi.com/trade-api/v2"
	}
	return "https://external-api.demo.kalshi.co/trade-api/v2"
}

// DefaultWSURL returns the margin WebSocket URL for env.
func DefaultWSURL(env string) string {
	if env == EnvProd {
		return "wss://external-api-margin-ws.kalshi.com/trade-api/ws/v2/margin"
	}
	return "wss://external-api-margin-ws.demo.kalshi.co/trade-api/ws/v2/margin"
}

// SampleYAML is written by `config init`.
func SampleYAML() string {
	return `# Kalshi Perp CLI configuration
# Prefer demo until you intentionally trade production.

env: demo
api_key: ""
private_key_path: ""
# base_url: ""   # optional override
# ws_url: ""     # optional override
format: table
timeout_sec: 30
`
}

// WriteSample creates the config file (and parent dirs) if missing.
func WriteSample(path string, force bool) error {
	if path == "" {
		path = DefaultConfigPath()
	}
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("config already exists: %s (use --force to overwrite)", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(SampleYAML()), 0o600)
}
