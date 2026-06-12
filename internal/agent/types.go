package agent

import "time"

type SessionState int

const (
	StateIdle SessionState = iota
	StateRunning
	StateError
	StateClosed
)

func (s SessionState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateError:
		return "error"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

type EventType int

const (
	EventInit EventType = iota
	EventText
	EventTextDone
	EventToolStart
	EventToolResult
	EventComplete
	EventError
)

type AgentEvent struct {
	Type    EventType
	Content string
	Tool    *ToolUse
	Meta    EventMeta
}

type ToolUse struct {
	ID    string
	Name  string
	Input string
	Output string
}

type EventMeta struct {
	Cost     float64
	Tokens   int
	Duration time.Duration
}
