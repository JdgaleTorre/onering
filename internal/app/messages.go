package app

import "github.com/josegale/lazycode/internal/agent"

type InputMode int

const (
	ModeNavigation InputMode = iota
	ModeInsert
)

type FocusPanel int

const (
	FocusSidebar FocusPanel = iota
	FocusMain
)

type AgentEventMsg struct {
	SessionID string
	Event     agent.AgentEvent
}

type SessionCreatedMsg struct {
	SessionID string
}

type SessionDeletedMsg struct {
	SessionID string
}

type ErrorMsg struct {
	Err error
}
