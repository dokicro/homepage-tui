package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HomepageURL     string        `yaml:"homepage_url"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	Auth            AuthConfig    `yaml:"auth"`
}

type AuthConfig struct {
	Username string            `yaml:"username"`
	Password string            `yaml:"password"`
	Headers  map[string]string `yaml:"headers"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	cfg.HomepageURL = strings.TrimRight(cfg.HomepageURL, "/")

	if cfg.HomepageURL == "" {
		return Config{}, fmt.Errorf("homepage_url is required")
	}

	if cfg.RefreshInterval == 0 {
		cfg.RefreshInterval = 30 * time.Second
	}

	return cfg, nil
}
