// File: internal/config/config.go
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Proxy struct {
		Target string `yaml:"target"`
	} `yaml:"proxy"`
	Cache struct {
		CacheType         string `yaml:"cache_type"`
		DefaultTTLSeconds int    `yaml:"default_ttl_seconds"`
		LRU               struct {
			Size int `yaml:"size"`
		} `yaml:"lru"`
	} `yaml:"cache"`
	Redis struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

func (c *Config) GetDefaultTTL() time.Duration {
	return time.Duration(c.Cache.DefaultTTLSeconds) * time.Second
}

func Load(path string) (*Config, error) {
	// ... (no changes to the Load function)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}