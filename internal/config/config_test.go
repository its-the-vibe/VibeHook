package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesRoutesAndDefaultsAddress(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := "routes:\n  /webhook/github: github-events\n  /webhook/slack: slack-events\n"

	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ListenAddress != ":8080" {
		t.Fatalf("expected default listen address ':8080', got %q", cfg.ListenAddress)
	}

	if got := cfg.Routes["/webhook/github"]; got != "github-events" {
		t.Fatalf("expected github-events channel, got %q", got)
	}
}

func TestLoadValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty routes",
			content: "routes: {}\n",
		},
		{
			name: "route missing leading slash",
			content: "routes:\n" +
				"  webhook/github: github-events\n",
		},
		{
			name: "empty channel",
			content: "routes:\n" +
				"  /webhook/github: \"\"\n",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tc.content), 0o600); err != nil {
				t.Fatalf("write config file: %v", err)
			}

			if _, err := Load(configPath); err == nil {
				t.Fatalf("expected Load to fail for case %q", tc.name)
			}
		})
	}
}
