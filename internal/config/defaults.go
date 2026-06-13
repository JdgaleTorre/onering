package config

func DefaultConfig() *Config {
	return &Config{
		Agents: map[string]AgentConfig{
			"claude": {
				Enabled: true,
				Command: "claude",
			},
			"codex": {
				Enabled: false,
				Command: "codex",
			},
			"aider": {
				Enabled: false,
				Command: "aider",
			},
			"opencode": {
				Enabled: true,
				Command: "opencode",
			},
		},
		SideApps: SideAppsConfig{
			Editor: "nvim .",
			Git:    "lazygit",
		},
		UI: UIConfig{
			SidebarWidth: 30,
			ShowCost:     true,
			ShowTokens:   true,
		},
	}
}
