package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/josegale/lazycode/internal/agent"
	"github.com/josegale/lazycode/internal/terminal"
)

type MainPanelModel struct {
	agentView   AgentViewModel
	prompt      PromptModel
	termViews   map[string]terminal.TermViewModel
	activeView  string
	focused     bool
	width       int
	height      int
	hasPTY      bool
	activeSess  agent.Session
	passthrough bool
	showInfo    bool
	projectInfo ProjectInfoModel
}

func NewMainPanelModel() MainPanelModel {
	return MainPanelModel{
		agentView:   NewAgentViewModel(),
		prompt:      NewPromptModel(),
		termViews:   make(map[string]terminal.TermViewModel),
		projectInfo: NewProjectInfoModel(),
	}
}

func (m MainPanelModel) SetSize(w, h int) MainPanelModel {
	m.width = w
	m.height = h
	innerW, innerH := m.termSize()
	m.agentView = m.agentView.SetSize(innerW, innerH-3)
	m.prompt = m.prompt.SetSize(innerW, 3)
	m.projectInfo = m.projectInfo.SetSize(innerW, innerH)
	for id, tv := range m.termViews {
		m.termViews[id] = tv.SetSize(m.termSize())
	}
	return m
}

// termSize is the panel's inner area: the border consumes one cell on
// each side. The pty and emulator must match it exactly.
func (m MainPanelModel) termSize() (int, int) {
	return m.width - 2, m.height - 2
}

func (m MainPanelModel) SetFocused(f bool) MainPanelModel {
	m.focused = f
	return m
}

func (m MainPanelModel) SetPassthrough(b bool) MainPanelModel {
	m.passthrough = b
	for id, tv := range m.termViews {
		m.termViews[id] = tv.SetPassthrough(b)
	}
	return m
}

func (m MainPanelModel) ShowInfo(show bool) MainPanelModel {
	m.showInfo = show
	return m
}

func (m MainPanelModel) SetKeyBindingGroups(groups []BindingGroup) MainPanelModel {
	m.projectInfo = m.projectInfo.SetKeyBindingGroups(groups)
	return m
}

func (m MainPanelModel) SetProjectName(name string) MainPanelModel {
	m.projectInfo = m.projectInfo.SetProjectName(name)
	return m
}

func (m MainPanelModel) SetSession(sess agent.Session) (MainPanelModel, tea.Cmd) {
	m.showInfo = false
	m.activeSess = sess
	if sess == nil {
		m.hasPTY = false
		m.activeView = ""
		return m, nil
	}
	if ptySess, ok := sess.(agent.PTYProvider); ok {
		m.hasPTY = true
		m.activeView = sess.ID()
		if _, exists := m.termViews[sess.ID()]; !exists {
			tv := terminal.NewTermViewModel(sess.ID(), ptySess.PTY())
			tv = tv.SetSize(m.termSize())
			tv = tv.SetPassthrough(m.passthrough)
			m.termViews[sess.ID()] = tv
			return m, tv.Init()
		}
	} else {
		m.hasPTY = false
		m.activeView = ""
	}
	return m, nil
}

func (m MainPanelModel) ActiveSession() agent.Session {
	return m.activeSess
}

func (m MainPanelModel) HasPTY() bool {
	return m.hasPTY
}

func (m MainPanelModel) RemoveSessionView(sessionID string) MainPanelModel {
	if tv, ok := m.termViews[sessionID]; ok {
		tv.Close()
		delete(m.termViews, sessionID)
	}
	if m.activeView == sessionID {
		m.activeView = ""
		m.hasPTY = false
	}
	return m
}

func (m MainPanelModel) ScrollTermView(direction int) (MainPanelModel, tea.Cmd) {
	if !m.hasPTY || m.activeView == "" {
		return m, nil
	}
	tv, ok := m.termViews[m.activeView]
	if !ok {
		return m, nil
	}
	_, h := m.termSize()
	half := h / 2
	if direction < 0 {
		tv = tv.ScrollUp(half)
	} else {
		tv = tv.ScrollDown(half)
	}
	m.termViews[m.activeView] = tv
	return m, nil
}

func (m MainPanelModel) Update(msg tea.Msg) (MainPanelModel, tea.Cmd) {
	var cmd tea.Cmd

	// PTY output is routed to its own view by session ID, no matter
	// which view is active, so background sessions keep reading.
	switch msg := msg.(type) {
	case terminal.OutputMsg:
		return m.updateTermView(msg.ID, msg)
	case terminal.TermErrorMsg:
		return m.updateTermView(msg.ID, msg)
	}

	if m.hasPTY && m.activeView != "" {
		return m.updateTermView(m.activeView, msg)
	}

	m.agentView, cmd = m.agentView.Update(msg)
	return m, cmd
}

func (m MainPanelModel) updateTermView(id string, msg tea.Msg) (MainPanelModel, tea.Cmd) {
	tv, ok := m.termViews[id]
	if !ok {
		return m, nil
	}
	updated, cmd := tv.Update(msg)
	m.termViews[id] = updated
	return m, cmd
}

func (m MainPanelModel) View() string {
	style := BorderNormal
	if m.focused {
		style = BorderFocused
	}

	var inner string

	if m.showInfo {
		inner = m.projectInfo.View()
	} else if m.hasPTY && m.activeView != "" {
		if tv, ok := m.termViews[m.activeView]; ok {
			inner = tv.View()
		}
	} else {
		inner = lipgloss.JoinVertical(
			lipgloss.Left,
			m.agentView.View(),
			m.prompt.View(),
		)
	}

	if inner == "" {
		inner = MutedStyle.Render("Start a session and send a prompt to begin.")
	}

	// Width/Height size the content area (the border adds one cell per
	// side), while MaxWidth/MaxHeight clip the final render including the
	// border — so they must be the full panel size or the border is cut.
	return style.
		Width(m.width - 2).
		MaxWidth(m.width).
		Height(m.height - 2).
		MaxHeight(m.height).
		Render(inner)
}
