package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddress string            `yaml:"listenAddress"`
	Routes        map[string]string `yaml:"routes"`
}

func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.ListenAddress == "" {
		cfg.ListenAddress = ":8080"
	}

	if len(cfg.Routes) == 0 {
		return Config{}, errors.New("routes must not be empty")
	}

	for routePath, channel := range cfg.Routes {
		if !strings.HasPrefix(routePath, "/") {
			return Config{}, fmt.Errorf("route %q must start with /", routePath)
		}
		if channel == "" {
			return Config{}, fmt.Errorf("route %q has empty channel", routePath)
		}
	}

	return cfg, nil
}
