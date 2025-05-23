package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	MongoDB struct {
		URI      string `yaml:"uri"`
		Database string `yaml:"database"`
	} `yaml:"mongodb"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"log"`
	Auth struct {
		Secret          string        `yaml:"secret"`
		AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
		RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
	} `yaml:"auth"`
}

func (c *Config) validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if c.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}
	if c.MongoDB.URI == "" {
		return fmt.Errorf("mongodb uri is required")
	}
	if c.MongoDB.Database == "" {
		return fmt.Errorf("mongodb database name is required")
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "json"
	}
	if c.Auth.Secret == "" {
		return fmt.Errorf("auth secret is required")
	}
	if c.Auth.AccessTokenTTL == 0 {
		return fmt.Errorf("access token TTL is required")
	}
	if c.Auth.RefreshTokenTTL == 0 {
		return fmt.Errorf("refresh token TTL is required")
	}
	return nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
