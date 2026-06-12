package agent

import (
	"context"
	"os/exec"
)

type CodexAgent struct {
	command string
}

func NewCodexAgent(command string) *CodexAgent {
	if command == "" {
		command = "codex"
	}
	return &CodexAgent{command: command}
}

func (a *CodexAgent) Name() string {
	return "codex"
}

func (a *CodexAgent) Available() (bool, error) {
	_, err := exec.LookPath(a.command)
	return err == nil, err
}

func (a *CodexAgent) StartSession(ctx context.Context, opts SessionOpts) (Session, error) {
	// TODO: implement Codex session
	return nil, nil
}

func (a *CodexAgent) ResumeSession(ctx context.Context, sessionID string) (Session, error) {
	// TODO: implement session resume
	return nil, nil
}
