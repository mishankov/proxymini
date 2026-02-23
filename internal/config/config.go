package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

func getStringOrDefault(name, def string) string {
	value := os.Getenv(name)

	if len(value) == 0 {
		return def
	}

	return strings.TrimSpace(value)
}

func getIntOrDefault(name string, def int) int {
	value := os.Getenv(name)
	if value == "" {
		return def
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return def
	}

	return result
}

type Config struct {
	Port       string
	ConfigPath string
	DBPath     string
	Retention  int
	Proxies    []Proxy `toml:"proxy"`
}

type Proxy struct {
	Prefix string `toml:"prefix"`
	Target string `toml:"target"`
}

func New() (*Config, error) {
	var config Config

	config.Port = getStringOrDefault("PROXYMINI_PORT", "14443")
	config.ConfigPath = getStringOrDefault("PROXYMINI_CONFIG", "proxymini.conf.toml")
	config.DBPath = getStringOrDefault("PROXYMINI_DB", "rl.db")
	config.Retention = getIntOrDefault("PROXYMINI_RETENTION", 0) // 0 means no retention (keep all logs)

	data, err := os.ReadFile(config.ConfigPath)
	if err != nil {
		return nil, err
	}

	_, err = toml.Decode(string(data), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) ReloadProxies() error {
	data, err := os.ReadFile(c.ConfigPath)
	if err != nil {
		return err
	}

	var freshConfig Config
	_, err = toml.Decode(string(data), &freshConfig)
	if err != nil {
		return err
	}

	c.Proxies = freshConfig.Proxies

	return nil
}
