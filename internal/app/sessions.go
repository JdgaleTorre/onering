package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JdgaleTorre/onering/internal/agent"
	"github.com/JdgaleTorre/onering/internal/ui"
)

func (m AppModel) activeSession() agent.Session {
	if m.activeIdx >= 0 && m.activeIdx < len(m.sessions) {
		return m.sessions[m.activeIdx]
	}
	return nil
}

// displayedSession is whatever the main panel currently shows: an app's
// session when one is active, the selected agent session otherwise.
func (m AppModel) displayedSession() agent.Session {
	if m.activeTask >= 0 && m.activeTask < len(m.tasks) && m.tasks[m.activeTask].IsPTY {
		return m.tasks[m.activeTask].Sess
	}
	if m.activeApp >= 0 && m.activeApp < len(m.apps) {
		return m.apps[m.activeApp].Sess
	}
	return m.activeSession()
}

func (m AppModel) addSession(sess agent.Session) (AppModel, tea.Cmd) {
	m.sessions = append(m.sessions, sess)
	m.activeIdx = len(m.sessions) - 1
	m.activeApp = -1
	m.showInfo = false
	m.cursorSec = ui.SectionSessions
	m.cursorIdx = m.activeIdx

	var cmd tea.Cmd
	if m.layout.LayoutMode() == ui.LayoutSplit {
		m.layout, cmd = m.layout.SetAgentSession(sess)
		m.focus = ui.FocusAgent
		m.layout = m.layout.SetFocus(ui.FocusAgent)
	} else {
		m.layout, cmd = m.layout.SetActiveSession(sess)
		m.focus = ui.FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
	}
	return m.syncSidebar(), cmd
}

func (m AppModel) removeSession(idx int) (AppModel, tea.Cmd) {
	if idx < 0 || idx >= len(m.sessions) {
		return m, nil
	}
	sessionID := m.sessions[idx].ID()
	m.sessions[idx].Close()
	m.sessions = append(m.sessions[:idx], m.sessions[idx+1:]...)
	if m.activeIdx >= len(m.sessions) {
		m.activeIdx = len(m.sessions) - 1
	}
	m.layout = m.layout.RemoveSessionView(sessionID)
	m.layout = m.layout.RemoveAgentSessionView(sessionID)
	var cmd tea.Cmd
	if m.activeApp < 0 {
		if m.layout.LayoutMode() == ui.LayoutSplit {
			m.layout, cmd = m.layout.SetAgentSession(m.activeSession())
		} else {
			m.layout, cmd = m.layout.SetActiveSession(m.activeSession())
		}
	}
	if len(m.sessions) == 0 && m.mode != ModePassthrough {
		m = m.exitToNavigation()
	}
	return m.syncSidebar(), cmd
}
