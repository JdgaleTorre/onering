package app

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/josegale/lazycode/internal/agent"
	"github.com/josegale/lazycode/internal/config"
	"github.com/josegale/lazycode/internal/terminal"
	"github.com/josegale/lazycode/internal/ui"
)

var sectionOrder = []ui.SidebarSection{ui.SectionSessions, ui.SectionApps}

type AppModel struct {
	config *config.Config
	keys   KeyMap
	mode   InputMode
	focus  FocusPanel

	agents   []agent.Agent
	sessions []agent.Session
	activeIdx int

	apps      []SideApp
	cursorSec ui.SidebarSection
	cursorIdx int
	// activeApp is the index of the app shown in the main panel;
	// -1 means a session (activeIdx) is shown instead.
	activeApp int

	projName  string
	gitBranch string
	showInfo  bool

	layout     ui.LayoutModel
	status     ui.StatusBarModel
	help       ui.HelpModel
	labelModal *LabelModal

	width  int
	height int
}

func New(cfg *config.Config) AppModel {
	reg := agent.NewDefaultRegistry(cfg)
	available := reg.Available()

	projName, gitBranch := readProjectInfo()

	m := AppModel{
		config:     cfg,
		keys:       DefaultKeyMap,
		mode:       ModeNavigation,
		focus:      FocusSidebar,
		agents:     available,
		sessions:   nil,
		activeIdx:  -1,
		apps:       buildSideApps(cfg),
		cursorSec:  ui.SectionSessions,
		cursorIdx:  0,
		activeApp:  -1,
		projName:   projName,
		gitBranch:  gitBranch,
		showInfo:   true,
		layout:     ui.NewLayoutModel(cfg),
		status:     ui.NewStatusBarModel(),
		help:       ui.NewHelpModel(DefaultKeyMap.NavigationBindings()),
		labelModal: NewLabelModal(available),
	}
	m.layout = m.layout.SetKeyBindingGroups(DefaultKeyMap.ImportantBindingGroups())
	m.layout = m.layout.SetProjectName(projName)
	m.layout = m.layout.ShowInfo(true)
	return m.syncSidebar()
}

func (m AppModel) activeSession() agent.Session {
	if m.activeIdx >= 0 && m.activeIdx < len(m.sessions) {
		return m.sessions[m.activeIdx]
	}
	return nil
}

// displayedSession is whatever the main panel currently shows: an app's
// session when one is active, the selected agent session otherwise.
func (m AppModel) displayedSession() agent.Session {
	if m.activeApp >= 0 && m.activeApp < len(m.apps) {
		return m.apps[m.activeApp].Sess
	}
	return m.activeSession()
}

func (m AppModel) sectionLen(s ui.SidebarSection) int {
	switch s {
	case ui.SectionSessions:
		return len(m.sessions)
	case ui.SectionApps:
		return len(m.apps)
	}
	return 0
}

func (m AppModel) appItems() []ui.AppItem {
	items := make([]ui.AppItem, len(m.apps))
	for i, a := range m.apps {
		items[i] = ui.AppItem{Name: a.Name, Running: a.Running()}
	}
	return items
}

func (m AppModel) sidebarData() ui.SidebarData {
	return ui.SidebarData{
		ProjectName:   m.projName,
		Branch:        m.gitBranch,
		Sessions:      m.sessions,
		Apps:          m.appItems(),
		CursorSection: m.cursorSec,
		CursorIdx:     m.cursorIdx,
	}
}

// syncSidebar clamps the cursor to a valid item and pushes the current
// state into the sidebar. Call it after any state change it displays.
func (m AppModel) syncSidebar() AppModel {
	if m.cursorSec == ui.SectionSessions {
		// Sessions is always focusable, even when empty.
		if n := m.sectionLen(ui.SectionSessions); n == 0 {
			m.cursorIdx = 0
		} else if m.cursorIdx >= n {
			m.cursorIdx = n - 1
		} else if m.cursorIdx < 0 {
			m.cursorIdx = 0
		}
	} else if m.sectionLen(m.cursorSec) == 0 {
		m.cursorSec = ui.SectionSessions
		m.cursorIdx = 0
	} else if m.cursorIdx >= m.sectionLen(m.cursorSec) {
		m.cursorIdx = m.sectionLen(m.cursorSec) - 1
	} else if m.cursorIdx < 0 {
		m.cursorIdx = 0
	}
	m.layout = m.layout.SetSidebar(m.sidebarData())
	return m
}

func (m AppModel) enterPassthrough() AppModel {
	m.mode = ModePassthrough
	m.status = m.status.SetMode("PASSTHROUGH")
	m.status = m.status.SetHints(" ctrl+u/d: scroll  ctrl+q: exit")
	m.help = m.help.SetBindings(m.keys.PassthroughBindings())
	m.layout = m.layout.SetPassthrough(true)
	return m
}

func (m AppModel) exitToNavigation() AppModel {
	m.mode = ModeNavigation
	m.status = m.status.SetMode("NORMAL")
	m.status = m.status.SetHints(" n: new session  e: editor  g: lazygit  q: quit")
	m.help = m.help.SetBindings(m.keys.NavigationBindings())
	m.layout = m.layout.SetPassthrough(false)
	return m
}

func (m AppModel) addSession(sess agent.Session) (AppModel, tea.Cmd) {
	m.sessions = append(m.sessions, sess)
	m.activeIdx = len(m.sessions) - 1
	m.activeApp = -1
	m.showInfo = false
	m.cursorSec = ui.SectionSessions
	m.cursorIdx = m.activeIdx
	layout, cmd := m.layout.SetActiveSession(sess)
	m.layout = layout
	m.focus = FocusMain
	m.layout = m.layout.SetFocus(ui.FocusMain)
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
	var cmd tea.Cmd
	if m.activeApp < 0 {
		layout, c := m.layout.SetActiveSession(m.activeSession())
		m.layout = layout
		cmd = c
	}
	if len(m.sessions) == 0 && m.mode != ModePassthrough {
		m = m.exitToNavigation()
	}
	return m.syncSidebar(), cmd
}

// moveCursor moves the sidebar cursor by delta (±1), flowing across
// sections in order: Sessions, Apps.
func (m AppModel) moveCursor(delta int) (AppModel, tea.Cmd) {
	sec, idx := m.cursorSec, m.cursorIdx
	target := idx + delta

	if target >= 0 && target < m.sectionLen(sec) {
		// Normal movement within section.
		idx = target
	} else if next, ok := m.adjacentSection(sec, delta); ok {
		// Move to adjacent section even if empty.
		sec = next

		if n := m.sectionLen(sec); n > 0 {
			if delta > 0 {
				idx = 0
			} else {
				idx = n - 1
			}
		} else {
			idx = -1 // no item selected in this section
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
	if sec == ui.SectionSessions && idx >= 0 {
		m.activeIdx = idx
		m.activeApp = -1
		layout, c := m.layout.SetActiveSession(m.sessions[idx])
		m.layout = layout
		cmd = c
	}

	return m.syncSidebar(), cmd
}

func (m AppModel) adjacentSection(from ui.SidebarSection, delta int) (ui.SidebarSection, bool) {
	pos := int(from)
	for {
		pos += delta
		if pos < 0 || pos >= len(sectionOrder) {
			return from, false
		}
		sec := sectionOrder[pos]
		if sec == ui.SectionSessions || m.sectionLen(sec) > 0 {
			return sec, true
		}
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		statusHeight := 1
		m.layout = m.layout.SetSize(msg.Width, msg.Height-statusHeight)
		m.status = m.status.SetWidth(msg.Width)
		m.help = m.help.SetSize(msg.Width, msg.Height)
		m.labelModal.SetSize(msg.Width, msg.Height)
		return m, nil

	case terminal.TermErrorMsg:
		return m.handleTermError(msg)

	case SessionLabelConfirmMsg:
		ag := m.agents[msg.AgentIdx]
		sess, err := ag.StartSession(context.Background(), agent.SessionOpts{})
		if err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("starting session: %w", err)}
			}
		}
		if sess == nil {
			return m, nil
		}
		if msg.Label != "" {
			sess.SetLabel(msg.Label)
		}
		var sessionCmd tea.Cmd
		m, sessionCmd = m.addSession(sess)
		if _, ok := sess.(agent.PTYProvider); ok {
			m = m.enterPassthrough()
		}
		return m, sessionCmd

	case tea.KeyMsg:
		if m.labelModal.Visible() {
			var modalCmd tea.Cmd
			m.labelModal, modalCmd = m.labelModal.Update(msg)
			return m, modalCmd
		}

		if m.help.Visible() {
			m.help, cmd = m.help.Update(msg)
			return m, cmd
		}

		switch m.mode {
		case ModePassthrough:
			return m.updatePassthroughMode(msg)
		case ModeInsert:
			return m.updateInsertMode(msg)
		default:
			return m.updateNavigationMode(msg)
		}
	}

	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

// handleTermError reacts to a side app's pty closing (the program exited
// on its own): clear its slot and fall back to the selected session.
func (m AppModel) handleTermError(msg terminal.TermErrorMsg) (tea.Model, tea.Cmd) {
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
		m.projName, m.gitBranch = readProjectInfo()
		m = m.syncSidebar()
		break
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

func (m AppModel) updateNavigationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "q":
		for _, s := range m.sessions {
			s.Close()
		}
		for _, a := range m.apps {
			if a.Sess != nil {
				a.Sess.Close()
			}
		}
		return m, tea.Quit

	case msg.String() == "?":
		m.help = m.help.Toggle()
		return m, nil

	case msg.String() == "0":
		m.showInfo = !m.showInfo
		m.layout = m.layout.ShowInfo(m.showInfo)
		return m, nil

	case msg.String() == "h":
		m.focus = FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		return m, nil

	case msg.String() == "l":
		m.focus = FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		return m, nil

	case msg.String() == "tab":
		if m.focus == FocusSidebar {
			m.focus = FocusMain
			m.layout = m.layout.SetFocus(ui.FocusMain)
		} else {
			m.focus = FocusSidebar
			m.layout = m.layout.SetFocus(ui.FocusSidebar)
		}
		return m, nil

	case msg.String() == "j":
		if m.focus == FocusSidebar {
			var moveCmd tea.Cmd
			m, moveCmd = m.moveCursor(1)
			return m, moveCmd
		}
		return m, nil

	case msg.String() == "k":
		if m.focus == FocusSidebar {
			var moveCmd tea.Cmd
			m, moveCmd = m.moveCursor(-1)
			return m, moveCmd
		}
		return m, nil

	case msg.String() == "enter":
		if m.focus == FocusSidebar {
			return m.activateCursor()
		}
		return m, nil

	case msg.String() == "e":
		return m.activateAppByName("editor")

	case msg.String() == "g":
		return m.activateAppByName("git")

	case msg.String() == "n":
		if len(m.agents) == 0 {
			return m, nil
		}
		m.labelModal.Open()
		return m, nil

	case msg.String() == "d":
		switch m.cursorSec {
		case ui.SectionSessions:
			var removeCmd tea.Cmd
			m, removeCmd = m.removeSession(m.cursorIdx)
			return m, removeCmd
		case ui.SectionApps:
			return m.killApp(m.cursorIdx)
		}
		return m, nil

	case msg.String() == "i":
		sess := m.displayedSession()
		if sess == nil {
			return m, nil
		}
		m.focus = FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		if _, ok := sess.(agent.PTYProvider); ok {
			m = m.enterPassthrough()
		} else {
			m.mode = ModeInsert
			m.status = m.status.SetMode("INSERT")
			m.status = m.status.SetHints(" esc: back")
			m.help = m.help.SetBindings(m.keys.InsertBindings())
		}
		return m, nil

	case msg.String() == "1":
		m.focus = FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		m.cursorSec = ui.SectionSessions
		m.cursorIdx = 0
		return m.syncSidebar(), nil

	case msg.String() == "2":
		m.focus = FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		m.cursorSec = ui.SectionApps
		m.cursorIdx = 0
		return m.syncSidebar(), nil

	case key.Matches(msg, m.keys.PageUp):
		var scrollCmd tea.Cmd
		m.layout, scrollCmd = m.layout.ScrollMainPanel(-1)
		return m, scrollCmd

	case key.Matches(msg, m.keys.PageDown):
		var scrollCmd tea.Cmd
		m.layout, scrollCmd = m.layout.ScrollMainPanel(1)
		return m, scrollCmd
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
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
		layout, cmd := m.layout.SetActiveSession(m.sessions[m.activeIdx])
		m.layout = layout
		m.focus = FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		if _, ok := m.sessions[m.activeIdx].(agent.PTYProvider); ok {
			m = m.enterPassthrough()
		}
		return m.syncSidebar(), cmd

	case ui.SectionApps:
		return m.activateApp(m.cursorIdx)
	}
	return m, nil
}

// activateApp launches the app if it is not running and shows it in the
// main panel in passthrough mode.
func (m AppModel) activateApp(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.apps) {
		return m, nil
	}
	if !m.apps[idx].Running() {
		sess, err := startSideApp(m.apps[idx])
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
	m.focus = FocusMain
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
	m.projName, m.gitBranch = readProjectInfo()
	return m.syncSidebar(), cmd
}

func (m AppModel) updateInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		return m.exitToNavigation(), nil
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

func (m AppModel) updatePassthroughMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.PassthroughEscape) {
		m = m.exitToNavigation()
		// Apps like lazygit may have changed the branch while embedded.
		m.projName, m.gitBranch = readProjectInfo()
		return m.syncSidebar(), nil
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	if m.width == 0 {
		return ""
	}

	if m.labelModal.Visible() {
		return m.labelModal.View()
	}

	view := m.layout.View() + "\n" + m.status.View()

	if m.help.Visible() {
		return m.help.View()
	}

	return view
}
