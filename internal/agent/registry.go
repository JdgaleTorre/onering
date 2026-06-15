package agent

import (
	"fmt"

	"github.com/josegale/onering/internal/config"
)

type Registry struct {
	agents map[string]Agent
}

func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]Agent),
	}
}

func NewDefaultRegistry(cfg *config.Config) *Registry {
	r := NewRegistry()

	commands := map[string]string{
		"claude":   "claude",
		"codex":    "codex",
		"aider":    "aider",
		"opencode": "opencode",
	}
	for name, defaultCmd := range commands {
		cmd := defaultCmd
		if a, ok := cfg.Agents[name]; ok {
			if a.Command != "" {
				cmd = a.Command
			}
		}
		switch name {
		case "claude":
			r.Register(NewClaudeAgent(cmd))
		case "codex":
			r.Register(NewCodexAgent(cmd))
		case "aider":
			r.Register(NewAiderAgent(cmd))
		case "opencode":
			r.Register(NewOpenCodeAgent(cmd))
		}
	}
	return r
}

func (r *Registry) Register(a Agent) {
	r.agents[a.Name()] = a
}

func (r *Registry) Get(name string) (Agent, error) {
	a, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q not registered", name)
	}
	return a, nil
}

func (r *Registry) Available() []Agent {
	var result []Agent
	for _, a := range r.agents {
		if ok, _ := a.Available(); ok {
			result = append(result, a)
		}
	}
	return result
}

func (r *Registry) All() []Agent {
	result := make([]Agent, 0, len(r.agents))
	for _, a := range r.agents {
		result = append(result, a)
	}
	return result
}
