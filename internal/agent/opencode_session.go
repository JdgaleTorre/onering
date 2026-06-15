package agent

import (
	"context"
	"os"
	"sync"

	"github.com/josegale/onering/internal/terminal"
)

type PTYSession struct {
	id        string
	label     string
	agentName string
	// exitInput is written to the pty on Close to let the program quit
	// cleanly; when empty the process is killed instead.
	exitInput string
	state     SessionState
	pty       *terminal.PTYHandle
	mu        sync.RWMutex
	closed    bool
}

func NewPTYSession(pty *terminal.PTYHandle, agentName, label, id, exitInput string) *PTYSession {
	return &PTYSession{
		id:        id,
		label:     label,
		agentName: agentName,
		exitInput: exitInput,
		state:     StateRunning,
		pty:       pty,
	}
}

func (s *PTYSession) ID() string {
	return s.id
}

func (s *PTYSession) Label() string {
	return s.label
}

func (s *PTYSession) SetLabel(l string) {
	s.label = l
}

func (s *PTYSession) AgentName() string {
	return s.agentName
}

func (s *PTYSession) Send(ctx context.Context, prompt string) error {
	_, err := s.pty.Write([]byte(prompt + "\n"))
	return err
}

func (s *PTYSession) Events() <-chan AgentEvent {
	return nil
}

func (s *PTYSession) PTY() *os.File {
	return s.pty.F
}

func (s *PTYSession) State() SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *PTYSession) Cancel() {
	s.Close()
}

func (s *PTYSession) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.state = StateClosed
	s.mu.Unlock()
	if s.exitInput == "" {
		return s.pty.Close()
	}
	return s.pty.GracefulClose(s.exitInput)
}
