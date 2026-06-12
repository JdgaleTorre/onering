package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainPanelModel struct {
	agentView AgentViewModel
	prompt    PromptModel
	focused   bool
	width     int
	height    int
}

func NewMainPanelModel() MainPanelModel {
	return MainPanelModel{
		agentView: NewAgentViewModel(),
		prompt:    NewPromptModel(),
	}
}

func (m MainPanelModel) SetSize(w, h int) MainPanelModel {
	m.width = w
	m.height = h

	promptH := 3
	viewH := h - promptH

	m.agentView = m.agentView.SetSize(w, viewH)
	m.prompt = m.prompt.SetSize(w, promptH)
	return m
}

func (m MainPanelModel) SetFocused(f bool) MainPanelModel {
	m.focused = f
	return m
}

func (m MainPanelModel) Update(msg tea.Msg) (MainPanelModel, tea.Cmd) {
	return m, nil
}

func (m MainPanelModel) View() string {
	style := BorderNormal
	if m.focused {
		style = BorderFocused
	}

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		m.agentView.View(),
		m.prompt.View(),
	)

	return style.
		Width(m.width - 2).
		Height(m.height - 2).
		Render(inner)
}
