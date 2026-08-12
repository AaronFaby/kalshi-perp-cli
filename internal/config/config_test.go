package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultsDemo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.yaml")
	cfg, err := Load(path, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != EnvDemo {
		t.Fatal(cfg.Env)
	}
	if cfg.BaseURL != DefaultBaseURL(EnvDemo) {
		t.Fatal(cfg.BaseURL)
	}
}

func TestLoad_FlagsWinOverFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("env: prod\napi_key: file-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, Config{Env: "demo", APIKey: "flag-key"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "demo" || cfg.APIKey != "flag-key" {
		t.Fatalf("%+v", cfg)
	}
}

func TestConfigIsProduction(t *testing.T) {
	if !(Config{Env: EnvProd, BaseURL: DefaultBaseURL(EnvDemo)}).IsProduction() {
		t.Fatal("env prod")
	}
	if (Config{Env: EnvDemo, BaseURL: DefaultBaseURL(EnvDemo)}).IsProduction() {
		t.Fatal("demo")
	}
	if !(Config{Env: EnvDemo, BaseURL: DefaultBaseURL(EnvProd)}).IsProduction() {
		t.Fatal("prod host")
	}
}

func TestWriteSample(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kalshi-perp", "config.yaml")
	if err := WriteSample(path, false); err != nil {
		t.Fatal(err)
	}
	if err := WriteSample(path, false); err == nil {
		t.Fatal("expected exists error")
	}
	if err := WriteSample(path, true); err != nil {
		t.Fatal(err)
	}
}
