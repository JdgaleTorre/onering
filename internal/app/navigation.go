package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JdgaleTorre/onering/internal/agent"
	"github.com/JdgaleTorre/onering/internal/ui"
)

var sectionOrder = []ui.SidebarSection{ui.SectionProjectInfo, ui.SectionSessions, ui.SectionApps, ui.SectionTasks}

func (m AppModel) sectionLen(s ui.SidebarSection) int {
	switch s {
	case ui.SectionProjectInfo:
		return 1
	case ui.SectionSessions:
		return len(m.sessions)
	case ui.SectionApps:
		return len(m.apps)
	case ui.SectionTasks:
		return len(m.tasks)
	}
	return 0
}

func (m AppModel) appItems() []ui.AppItem {
	items := make([]ui.AppItem, len(m.apps))
	for i, a := range m.apps {
		items[i] = ui.AppItem{Name: a.Name, Running: a.Running(), Installed: a.Installed}
	}
	return items
}

func (m AppModel) sidebarData() ui.SidebarData {
	return ui.SidebarData{
		ProjectName:   m.projName,
		Branch:        m.gitBranch,
		Sessions:      m.sessions,
		Apps:          m.appItems(),
		Tasks:         m.taskItems(),
		CursorSection: m.cursorSec,
		CursorIdx:     m.cursorIdx,
	}
}

// syncSidebar clamps the cursor to a valid item and pushes the current
// state into the sidebar. Call it after any state change it displays.
func (m AppModel) syncSidebar() AppModel {
	if m.cursorSec == ui.SectionProjectInfo {
		m.cursorIdx = 0
	} else if m.cursorSec == ui.SectionSessions {
		if n := m.sectionLen(ui.SectionSessions); n == 0 {
			m.cursorIdx = 0
		} else if m.cursorIdx >= n {
			m.cursorIdx = n - 1
		} else if m.cursorIdx < 0 {
			m.cursorIdx = 0
		}
	} else if m.sectionLen(m.cursorSec) == 0 {
		m.cursorIdx = 0
	} else if m.cursorIdx >= m.sectionLen(m.cursorSec) {
		m.cursorIdx = m.sectionLen(m.cursorSec) - 1
	} else if m.cursorIdx < 0 {
		m.cursorIdx = 0
	}
	if m.cursorSec == ui.SectionTasks && m.cursorIdx >= 0 && m.cursorIdx < len(m.tasks) {
		t := m.tasks[m.cursorIdx]
		dirPart := ""
		if t.Dir != "" {
			dirPart = t.Dir + "/"
		}
		info := fmt.Sprintf("%s%s → %s", dirPart, t.Name, t.Command)
		m.status = m.status.SetTaskInfo(info)
	} else {
		m.status = m.status.ClearTaskInfo()
	}

	m.layout = m.layout.SetSidebar(m.sidebarData())
	if m.mode == ModeNavigation {
		m.status = m.status.SetHints(m.navigationHints())
	}
	return m
}

// moveCursor moves the sidebar cursor by delta (±1), flowing across
// sections in order: Sessions, Apps.
func (m AppModel) moveCursor(delta int) (AppModel, tea.Cmd) {
	sec, idx := m.cursorSec, m.cursorIdx
	target := idx + delta

	if target >= 0 && target < m.sectionLen(sec) {
		idx = target
	} else if next, ok := m.adjacentSection(sec, delta); ok {
		sec = next

		if n := m.sectionLen(sec); n > 0 {
			if delta > 0 {
				idx = 0
			} else {
				idx = n - 1
			}
		} else {
			idx = 0
		}
	} else if m.sectionLen(sec) > 0 {
		if target < 0 {
			idx = 0
		} else {
			idx = m.sectionLen(sec) - 1
		}
	} else {
		return m.syncSidebar(), nil
	}

	m.cursorSec, m.cursorIdx = sec, idx

	var cmd tea.Cmd
	switch {
	case sec == ui.SectionProjectInfo:
		m.showInfo = true
		m.layout = m.layout.ShowInfo(true)
	case sec == ui.SectionSessions && idx >= 0 && idx < len(m.sessions):
		m.showInfo = false
		m.activeIdx = idx
		m.activeApp = -1
		layout, c := m.layout.SetActiveSession(m.sessions[idx])
		m.layout = layout
		cmd = c
	case sec == ui.SectionSessions:
		m.showInfo = true
		m.layout = m.layout.ShowInfo(true)
	}

	return m.syncSidebar(), cmd
}

func (m AppModel) adjacentSection(from ui.SidebarSection, delta int) (ui.SidebarSection, bool) {
	pos := int(from) + delta
	if pos < 0 || pos >= len(sectionOrder) {
		return from, false
	}
	return sectionOrder[pos], true
}

// activateCursor performs Enter on the sidebar item under the cursor.
func (m AppModel) activateCursor() (tea.Model, tea.Cmd) {
	switch m.cursorSec {
	case ui.SectionSessions:
		if m.cursorIdx >= len(m.sessions) {
			return m, nil
		}
		m.activeIdx = m.cursorIdx
		m.activeApp = -1
		m.showInfo = false
		var cmd tea.Cmd
		if m.layout.LayoutMode() == ui.LayoutSplit {
			m.layout, cmd = m.layout.SetAgentSession(m.sessions[m.activeIdx])
			m.focus = ui.FocusAgent
			m.layout = m.layout.SetFocus(ui.FocusAgent)
		} else {
			m.layout, cmd = m.layout.SetActiveSession(m.sessions[m.activeIdx])
			m.focus = ui.FocusMain
			m.layout = m.layout.SetFocus(ui.FocusMain)
		}
		if _, ok := m.sessions[m.activeIdx].(agent.PTYProvider); ok {
			m = m.enterPassthrough()
		}
		return m.syncSidebar(), cmd

	case ui.SectionProjectInfo:
		m.projectModal.Open(m.state.RecentProjects)
		return m, nil

	case ui.SectionApps:
		return m.activateApp(m.cursorIdx)

	case ui.SectionTasks:
		return m.activateTask(m.cursorIdx, false)
	}
	return m, nil
}
