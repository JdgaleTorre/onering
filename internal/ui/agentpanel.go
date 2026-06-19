package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JdgaleTorre/onering/internal/agent"
	"github.com/JdgaleTorre/onering/internal/terminal"
)

type AgentPanelModel struct {
	termViews   map[string]terminal.TermViewModel
	activeView  string
	focused     bool
	width       int
	height      int
	passthrough bool
}

func NewAgentPanelModel() AgentPanelModel {
	return AgentPanelModel{
		termViews: make(map[string]terminal.TermViewModel),
	}
}

func (m AgentPanelModel) SetSize(w, h int) AgentPanelModel {
	m.width = w
	m.height = h
	for id, tv := range m.termViews {
		m.termViews[id] = tv.SetSize(m.termSize())
	}
	return m
}

func (m AgentPanelModel) termSize() (int, int) {
	return m.width - 2, m.height - 2
}

func (m AgentPanelModel) SetFocused(f bool) AgentPanelModel {
	m.focused = f
	return m
}

func (m AgentPanelModel) SetPassthrough(b bool) AgentPanelModel {
	m.passthrough = b
	for id, tv := range m.termViews {
		m.termViews[id] = tv.SetPassthrough(b)
	}
	return m
}

func (m AgentPanelModel) SetSession(sess agent.Session) (AgentPanelModel, tea.Cmd) {
	if sess == nil {
		m.activeView = ""
		return m, nil
	}
	ptySess, ok := sess.(agent.PTYProvider)
	if !ok {
		m.activeView = ""
		return m, nil
	}
	m.activeView = sess.ID()
	if _, exists := m.termViews[sess.ID()]; !exists {
		tv := terminal.NewTermViewModel(sess.ID(), ptySess.PTY())
		tv = tv.SetSize(m.termSize())
		tv = tv.SetPassthrough(m.passthrough)
		m.termViews[sess.ID()] = tv
		return m, tv.Init()
	}
	return m, nil
}

func (m AgentPanelModel) RemoveSessionView(sessionID string) AgentPanelModel {
	if tv, ok := m.termViews[sessionID]; ok {
		tv.Close()
		delete(m.termViews, sessionID)
	}
	if m.activeView == sessionID {
		m.activeView = ""
	}
	return m
}

func (m AgentPanelModel) AttachTermView(id string, tv terminal.TermViewModel) AgentPanelModel {
	tv = tv.SetSize(m.termSize())
	tv = tv.SetPassthrough(m.passthrough)
	m.termViews[id] = tv
	m.activeView = id
	return m
}

func (m AgentPanelModel) DetachTermView(id string) (AgentPanelModel, terminal.TermViewModel, bool) {
	tv, ok := m.termViews[id]
	if !ok {
		return m, terminal.TermViewModel{}, false
	}
	delete(m.termViews, id)
	if m.activeView == id {
		m.activeView = ""
	}
	return m, tv, true
}

func (m AgentPanelModel) HasActiveView() bool {
	return m.activeView != ""
}

func (m AgentPanelModel) ActiveViewID() string {
	return m.activeView
}

func (m AgentPanelModel) ScrollTermView(direction int) (AgentPanelModel, tea.Cmd) {
	if m.activeView == "" {
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

func (m AgentPanelModel) Update(msg tea.Msg) (AgentPanelModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case terminal.OutputMsg:
		return m.updateTermView(msg.ID, msg)
	case terminal.TermErrorMsg:
		return m.updateTermView(msg.ID, msg)
	case terminal.ClearToastMsg:
		return m.updateTermView(msg.ID, msg)
	case terminal.ColorSchemeChangedMsg:
		for id, tv := range m.termViews {
			m.termViews[id] = tv.UpdateHostColors()
		}
		return m, nil
	}

	if m.activeView != "" {
		return m.updateTermView(m.activeView, msg)
	}

	return m, cmd
}

func (m AgentPanelModel) updateTermView(id string, msg tea.Msg) (AgentPanelModel, tea.Cmd) {
	tv, ok := m.termViews[id]
	if !ok {
		return m, nil
	}
	updated, cmd := tv.Update(msg)
	m.termViews[id] = updated
	return m, cmd
}

func (m AgentPanelModel) View() string {
	style := BorderNormal
	if m.focused {
		style = BorderFocused
	}

	var inner string
	if m.activeView != "" {
		if tv, ok := m.termViews[m.activeView]; ok {
			inner = tv.View()
		}
	}

	if inner == "" {
		inner = MutedStyle.Render("No active session")
	}

	return style.
		Width(m.width - 2).
		MaxWidth(m.width).
		Height(m.height - 2).
		MaxHeight(m.height).
		Render(inner)
}
