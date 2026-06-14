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
	Tasks    TasksConfig            `yaml:"tasks"`
}

type TasksConfig struct {
	PackageManager string `yaml:"package_manager,omitempty"`
}

type AgentConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Command   string   `yaml:"command"`
	Model     string   `yaml:"model,omitempty"`
	ExtraArgs []string `yaml:"extra_args,omitempty"`
}

type SideAppsConfig struct {
	Editor string          `yaml:"editor"`
	Git    string          `yaml:"git"`
	Docker string          `yaml:"docker"`
	Extra  []ExtraAppDef   `yaml:"extra"`
	Enable map[string]bool `yaml:"enable"`
}

type ExtraAppDef struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
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

func ConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "lazycode")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazycode")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}
