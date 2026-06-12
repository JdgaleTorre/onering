package agent

import (
	"context"
	"os/exec"
)

type AiderAgent struct {
	command string
}

func NewAiderAgent(command string) *AiderAgent {
	if command == "" {
		command = "aider"
	}
	return &AiderAgent{command: command}
}

func (a *AiderAgent) Name() string {
	return "aider"
}

func (a *AiderAgent) Available() (bool, error) {
	_, err := exec.LookPath(a.command)
	return err == nil, err
}

func (a *AiderAgent) StartSession(ctx context.Context, opts SessionOpts) (Session, error) {
	// TODO: implement Aider session
	return nil, nil
}

func (a *AiderAgent) ResumeSession(ctx context.Context, sessionID string) (Session, error) {
	// TODO: implement session resume
	return nil, nil
}
