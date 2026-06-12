package agent

import (
	"context"
	"os/exec"
)

type ClaudeAgent struct {
	command string
}

func NewClaudeAgent(command string) *ClaudeAgent {
	if command == "" {
		command = "claude"
	}
	return &ClaudeAgent{command: command}
}

func (a *ClaudeAgent) Name() string {
	return "claude"
}

func (a *ClaudeAgent) Available() (bool, error) {
	_, err := exec.LookPath(a.command)
	return err == nil, err
}

func (a *ClaudeAgent) StartSession(ctx context.Context, opts SessionOpts) (Session, error) {
	// TODO: implement Claude Code stream-json session
	return nil, nil
}

func (a *ClaudeAgent) ResumeSession(ctx context.Context, sessionID string) (Session, error) {
	// TODO: implement session resume
	return nil, nil
}
