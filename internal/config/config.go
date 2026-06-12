package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agents   map[string]AgentConfig `yaml:"agents"`
	SideApps SideAppsConfig         `yaml:"side_apps"`
	UI       UIConfig               `yaml:"ui"`
}

type AgentConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Command  string `yaml:"command"`
	Model    string `yaml:"model,omitempty"`
	ExtraArgs []string `yaml:"extra_args,omitempty"`
}

type SideAppsConfig struct {
	Editor string `yaml:"editor"`
	Git    string `yaml:"git"`
}

type UIConfig struct {
	SidebarWidth int  `yaml:"sidebar_width"`
	ShowCost     bool `yaml:"show_cost"`
	ShowTokens   bool `yaml:"show_tokens"`
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func configPath() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "lazycode", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazycode", "config.yaml")
}
