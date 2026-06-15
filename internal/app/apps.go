package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/josegale/onering/internal/terminal"
	"github.com/josegale/onering/internal/ui"
)

// activateApp launches the app if it is not running and shows it in the
// main panel in passthrough mode.
func (m AppModel) activateApp(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.apps) {
		return m, nil
	}
	if !m.apps[idx].Installed {
		m.activeApp = idx
		m.layout = m.layout.SetInfoText(installHint(m.apps[idx].Name, m.apps[idx].Cmd))
		m.focus = ui.FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		return m.syncSidebar(), nil
	}
	if !m.apps[idx].Running() {
		sess, err := startSideApp(m.apps[idx], m.projectDir)
		if err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("starting %s: %w", m.apps[idx].Name, err)}
			}
		}
		m.apps[idx].Sess = sess
	}
	m.activeApp = idx
	m.cursorSec = ui.SectionApps
	m.cursorIdx = idx
	layout, cmd := m.layout.SetActiveSession(m.apps[idx].Sess)
	m.layout = layout
	m.focus = ui.FocusMain
	m.layout = m.layout.SetFocus(ui.FocusMain)
	m = m.enterPassthrough()
	return m.syncSidebar(), cmd
}

func (m AppModel) activateAppByName(name string) (tea.Model, tea.Cmd) {
	for i, a := range m.apps {
		if a.Name == name {
			return m.activateApp(i)
		}
	}
	return m, nil
}

// killApp force-stops a running app; the slot stays listed as stopped.
func (m AppModel) killApp(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.apps) || m.apps[idx].Sess == nil {
		return m, nil
	}
	id := m.apps[idx].Sess.ID()
	m.apps[idx].Sess.Close()
	m.apps[idx].Sess = nil
	m.layout = m.layout.RemoveSessionView(id)
	var cmd tea.Cmd
	if m.activeApp == idx {
		m.activeApp = -1
		layout, c := m.layout.SetActiveSession(m.activeSession())
		m.layout = layout
		cmd = c
	}
	m.projName, m.gitBranch = readProjectInfo(m.projectDir)
	return m.syncSidebar(), cmd
}

// handleTermError reacts to a side app's pty closing (the program exited
// on its own): clear its slot and fall back to the selected session.
func (m AppModel) handleTermError(msg terminal.TermErrorMsg) (tea.Model, tea.Cmd) {
	if updated, handled := m.handleTaskTermError(msg); handled {
		m = updated
		return m.syncSidebar(), nil
	}

	for i := range m.apps {
		if m.apps[i].Sess == nil || m.apps[i].Sess.ID() != msg.ID {
			continue
		}
		m.apps[i].Sess.Close()
		m.apps[i].Sess = nil
		m.layout = m.layout.RemoveSessionView(msg.ID)
		if m.activeApp == i {
			m.activeApp = -1
			if m.mode == ModePassthrough {
				m = m.exitToNavigation()
			}
			layout, _ := m.layout.SetActiveSession(m.activeSession())
			m.layout = layout
		}
		m.projName, m.gitBranch = readProjectInfo(m.projectDir)
		m = m.syncSidebar()
		break
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}
