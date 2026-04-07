package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`
homepage_url: http://homepage.local:3000
refresh_interval: 45s
auth:
  username: admin
  password: secret
  headers:
    X-Api-Key: abc123
`), 0644)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HomepageURL != "http://homepage.local:3000" {
		t.Errorf("got URL %q, want %q", cfg.HomepageURL, "http://homepage.local:3000")
	}
	if cfg.RefreshInterval != 45*time.Second {
		t.Errorf("got interval %v, want %v", cfg.RefreshInterval, 45*time.Second)
	}
	if cfg.Auth.Username != "admin" {
		t.Errorf("got username %q, want %q", cfg.Auth.Username, "admin")
	}
	if cfg.Auth.Password != "secret" {
		t.Errorf("got password %q, want %q", cfg.Auth.Password, "secret")
	}
	if cfg.Auth.Headers["X-Api-Key"] != "abc123" {
		t.Errorf("got header %q, want %q", cfg.Auth.Headers["X-Api-Key"], "abc123")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`
homepage_url: http://localhost:3000
`), 0644)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RefreshInterval != 30*time.Second {
		t.Errorf("got interval %v, want %v", cfg.RefreshInterval, 30*time.Second)
	}
}

func TestLoadConfig_MissingURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`
refresh_interval: 10s
`), 0644)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for missing homepage_url")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
