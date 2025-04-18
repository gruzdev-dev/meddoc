package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
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
	return nil
}

// Load loads configuration from file and environment variables
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Override with environment variables if set
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		cfg.MongoDB.URI = uri
	}
	if db := os.Getenv("MONGODB_DATABASE"); db != "" {
		cfg.MongoDB.Database = db
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Log.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Log.Format = format
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
