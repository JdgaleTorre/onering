package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/josegale/lazycode/internal/agent"
	"github.com/josegale/lazycode/internal/ui"
)

type LabelModal struct {
	visible        bool
	value          strings.Builder
	agents         []agent.Agent
	agentIdx       int
	defaultAgentIdx int
	width          int
	height         int
}

func NewLabelModal(agents []agent.Agent) *LabelModal {
	defaultIdx := 0
	for i, a := range agents {
		if a.Name() == "opencode" {
			defaultIdx = i
			break
		}
	}
	return &LabelModal{
		agents:          agents,
		agentIdx:        defaultIdx,
		defaultAgentIdx: defaultIdx,
	}
}

func (m *LabelModal) Open() {
	m.visible = true
	m.value.Reset()
	m.agentIdx = m.defaultAgentIdx
}

func (m *LabelModal) Visible() bool {
	return m.visible
}

func (m *LabelModal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *LabelModal) Update(msg tea.Msg) (*LabelModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "esc":
			m.visible = false
			return m, nil
		case msg.String() == "enter":
			label := m.value.String()
			idx := m.agentIdx
			m.visible = false
			return m, func() tea.Msg {
				return SessionLabelConfirmMsg{
					Label:    label,
					AgentIdx: idx,
				}
			}
		case msg.String() == "up":
			if m.agentIdx > 0 {
				m.agentIdx--
			}
		case msg.String() == "down":
			if m.agentIdx < len(m.agents)-1 {
				m.agentIdx++
			}
		case msg.String() == "backspace" || msg.String() == "ctrl+backspace":
			str := m.value.String()
			if len(str) > 0 {
				m.value.Reset()
				m.value.WriteString(str[:len(str)-1])
			}
		default:
			if len(msg.String()) == 1 && msg.String()[0] >= 32 {
				m.value.WriteByte(msg.String()[0])
			}
		}
	}

	return m, nil
}

func (m *LabelModal) View() string {
	if !m.visible {
		return ""
	}

	title := ui.TitleStyle.Render("New Session")

	var agentLines []string
	for i, a := range m.agents {
		name := a.Name()
		lineStyle := lipgloss.NewStyle().Foreground(ui.ColorText)
		prefix := "  "
		if i == m.agentIdx {
			prefix = "▸ "
			lineStyle = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
		}
		agentLines = append(agentLines, lineStyle.Render(prefix+name))
	}

	input := m.value.String()
	if input == "" {
		input = ui.MutedStyle.Render("type a label...")
	}
	cursor := lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Render("█")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		"Agent:",
		strings.Join(agentLines, "\n"),
		"",
		"Label (optional):",
		input+cursor,
		"",
		ui.MutedStyle.Render("↑↓ select agent  Enter confirm  Esc cancel"),
	)

	box := ui.BorderFocused.
		Width(44).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}
