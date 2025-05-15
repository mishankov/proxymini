package config

import (
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

type Config struct {
	Port       string
	ConfigPath string
	DBPath     string
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
