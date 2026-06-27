package config

func DefaultConfig() *Config {
	return &Config{
		Agents: map[string]AgentConfig{
			"claude": {
				Enabled: true,
				Command: "claude",
			},
			"codex": {
				Enabled: true,
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
			Docker: "lazydocker",
			Enable: map[string]bool{},
		},
		UI: UIConfig{
			SidebarWidth: 30,
			ShowCost:     true,
			ShowTokens:   true,
		},
	}
}
