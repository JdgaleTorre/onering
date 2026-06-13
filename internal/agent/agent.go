package agent

import (
	"context"
	"os"
)

type Agent interface {
	Name() string
	Available() (bool, error)
	StartSession(ctx context.Context, opts SessionOpts) (Session, error)
	ResumeSession(ctx context.Context, sessionID string) (Session, error)
}

type Session interface {
	ID() string
	Label() string
	SetLabel(string)
	AgentName() string
	Send(ctx context.Context, prompt string) error
	Events() <-chan AgentEvent
	Cancel()
	Close() error
	State() SessionState
}

type PTYProvider interface {
	PTY() *os.File
}

type SessionOpts struct {
	Model     string
	WorkDir   string
	ExtraArgs []string
}
