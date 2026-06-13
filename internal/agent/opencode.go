package agent

import (
	"context"
	"os/exec"

	"github.com/josegale/lazycode/internal/terminal"
	"github.com/josegale/lazycode/internal/util"
)

type OpenCodeAgent struct {
	command string
}

func NewOpenCodeAgent(command string) *OpenCodeAgent {
	if command == "" {
		command = "opencode"
	}
	return &OpenCodeAgent{command: command}
}

func (a *OpenCodeAgent) Name() string {
	return "opencode"
}

func (a *OpenCodeAgent) Available() (bool, error) {
	_, err := exec.LookPath(a.command)
	return err == nil, err
}

func (a *OpenCodeAgent) StartSession(ctx context.Context, opts SessionOpts) (Session, error) {
	cmd := exec.CommandContext(ctx, a.command)
	cmd.Dir = opts.WorkDir

	ptyHandle, err := terminal.StartPTY(cmd)
	if err != nil {
		return nil, err
	}

	return NewPTYSession(ptyHandle, "opencode", "opencode", util.NewID(), "/exit"), nil
}

func (a *OpenCodeAgent) ResumeSession(ctx context.Context, sessionID string) (Session, error) {
	return nil, nil
}
