package agent

import "context"

type Agent interface {
	Name() string
	Available() (bool, error)
	StartSession(ctx context.Context, opts SessionOpts) (Session, error)
	ResumeSession(ctx context.Context, sessionID string) (Session, error)
}

type Session interface {
	ID() string
	Send(ctx context.Context, prompt string) error
	Events() <-chan AgentEvent
	Cancel()
	Close() error
	State() SessionState
}

type SessionOpts struct {
	Model    string
	WorkDir  string
	ExtraArgs []string
}
