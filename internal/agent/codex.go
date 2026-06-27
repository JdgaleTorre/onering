package agent

import "context"

type CodexAgent struct {
	BaseAgent
}

func NewCodexAgent(command string) *CodexAgent {
	if command == "" {
		command = "codex"
	}
	return &CodexAgent{BaseAgent{name: "codex", command: command}}
}

func (a *CodexAgent) StartSession(ctx context.Context, opts SessionOpts) (Session, error) {
	return a.StartPTYSession(ctx, opts, "")
}
