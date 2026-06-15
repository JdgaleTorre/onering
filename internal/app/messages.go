package app

import (
	"github.com/josegale/onering/internal/agent"
)

type InputMode int

const (
	ModeNavigation InputMode = iota
	ModeInsert
	ModePassthrough
)

type AgentEventMsg struct {
	SessionID string
	Event     agent.AgentEvent
}

type SessionCreatedMsg struct {
	SessionID string
	Label     string
}

type SessionDeletedMsg struct {
	SessionID string
}

type SessionLabelConfirmMsg struct {
	Label    string
	AgentIdx int
}

type PASSTHROUGHMsg struct {
	Data []byte
}

type ErrorMsg struct {
	Err error
}
