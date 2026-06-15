package agent

import (
	"context"
	"os/exec"

	"github.com/josegale/onering/internal/terminal"
	"github.com/josegale/onering/internal/util"
)

type BaseAgent struct {
	name    string
	command string
}

func (a *BaseAgent) Name() string {
	return a.name
}

func (a *BaseAgent) Available() (bool, error) {
	_, err := exec.LookPath(a.command)
	return err == nil, err
}

func (a *BaseAgent) ResumeSession(ctx context.Context, sessionID string) (Session, error) {
	return nil, nil
}

func (a *BaseAgent) StartPTYSession(ctx context.Context, opts SessionOpts, exitInput string) (Session, error) {
	cmd := exec.CommandContext(ctx, a.command)
	cmd.Dir = opts.WorkDir

	ptyHandle, err := terminal.StartPTY(cmd)
	if err != nil {
		return nil, err
	}

	return NewPTYSession(ptyHandle, a.name, a.name, util.NewID(), exitInput), nil
}
