package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddress     string            `yaml:"listenAddress"`
	RedisAddress      string            `yaml:"redisAddress"`
	RedisDB           int               `yaml:"redisDB"`
	BasicAuthEnabled  bool              `yaml:"basicAuthEnabled"`
	BasicAuthUsername string            `yaml:"basicAuthUsername"`
	Routes            map[string]string `yaml:"routes"`
}

func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	cfg := Config{
		ListenAddress:     ":8080",
		RedisAddress:      "localhost:6379",
		BasicAuthEnabled:  true,
		BasicAuthUsername: "vibehook",
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.ListenAddress == "" {
		return Config{}, errors.New("listenAddress must not be empty")
	}

	if cfg.RedisAddress == "" {
		return Config{}, errors.New("redisAddress must not be empty")
	}

	if cfg.RedisDB < 0 {
		return Config{}, errors.New("redisDB must be greater than or equal to 0")
	}

	if cfg.BasicAuthEnabled && cfg.BasicAuthUsername == "" {
		return Config{}, errors.New("basicAuthUsername must not be empty when basicAuthEnabled is true")
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
