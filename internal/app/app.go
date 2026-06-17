package app

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/JdgaleTorre/onering/internal/agent"
	"github.com/JdgaleTorre/onering/internal/config"
	"github.com/JdgaleTorre/onering/internal/task"
	"github.com/JdgaleTorre/onering/internal/terminal"
	"github.com/JdgaleTorre/onering/internal/ui"
)

const (
	minWidth       = 30
	minHeight      = 8
	collapseWidth  = 55
)

type AppModel struct {
	config *config.Config
	keys   KeyMap
	mode   InputMode
	focus  ui.FocusPanel

	agents    []agent.Agent
	sessions  []agent.Session
	activeIdx int

	apps      []SideApp
	tasks     []TaskRun
	cursorSec ui.SidebarSection
	cursorIdx int
	// activeApp is the index of the app shown in the main panel;
	// -1 means a session (activeIdx) is shown instead.
	activeApp  int
	activeTask int

	projName   string
	gitBranch  string
	projectDir string
	showInfo   bool

	layout       ui.LayoutModel
	status       ui.StatusBarModel
	help         ui.HelpModel
	labelModal   *LabelModal
	projectModal ui.ProjectModal
	state        *config.State

	width  int
	height int

	sidebarCollapsed bool
	tooSmall         bool

	quitting      bool
	shutdownItems []shutdownItem
	shutdownFrame int
}

func New(cfg *config.Config, version string) AppModel {
	reg := agent.NewDefaultRegistry(cfg)
	available := reg.Available()

	wd, _ := os.Getwd()
	state := config.LoadState()
	projName, gitBranch := readProjectInfo(wd)

	if gitBranch != "" {
		state.RecordProject(wd)
		state.Save()
	}

	var taskRuns []TaskRun
	if stored := state.LoadProjectTasks(wd); len(stored) > 0 {
		raw := task.TasksFromStored(stored)
		sorted := task.SortWithPreferred(raw, state.PreferredTasks[wd])
		taskRuns = make([]TaskRun, len(sorted))
		for i, t := range sorted {
			taskRuns[i] = TaskRun{Task: t, Status: ui.TaskPending}
		}
	} else {
		scannedTasks := task.ScanTasks(wd, cfg.Tasks.PackageManager)
		state.SaveProjectTasks(wd, task.TasksToStored(scannedTasks))
		sortedTasks := task.SortWithPreferred(scannedTasks, state.PreferredTasks[wd])
		taskRuns = make([]TaskRun, len(sortedTasks))
		for i, t := range sortedTasks {
			taskRuns[i] = TaskRun{Task: t, Status: ui.TaskPending}
		}
	}

	m := AppModel{
		config:       cfg,
		keys:         DefaultKeyMap,
		mode:         ModeNavigation,
		focus:        ui.FocusSidebar,
		agents:       available,
		sessions:     nil,
		activeIdx:    -1,
		apps:         buildSideApps(cfg),
		tasks:        taskRuns,
		cursorSec:    ui.SectionProjectInfo,
		cursorIdx:    0,
		activeApp:    -1,
		activeTask:   -1,
		projName:     projName,
		gitBranch:    gitBranch,
		projectDir:   wd,
		showInfo:     true,
		layout:       ui.NewLayoutModel(cfg),
		status:       ui.NewStatusBarModel().SetVersion(version),
		help:         ui.NewHelpModel(DefaultKeyMap.NavigationBindings()),
		labelModal:   NewLabelModal(available),
		projectModal: ui.NewProjectModal(),
		state:        state,
	}
	m.layout = m.layout.SetKeyBindingGroups(DefaultKeyMap.ImportantBindingGroups())
	m.layout = m.layout.SetProjectName(projName)
	m.layout = m.layout.ShowInfo(true)
	if gitBranch == "" {
		m.projectModal.Open(m.state.RecentProjects)
	}
	return m.syncSidebar()
}

func (m AppModel) enterPassthrough() AppModel {
	m.mode = ModePassthrough
	m.status = m.status.SetMode("PASSTHROUGH")
	m.status = m.status.SetHints(" ctrl+j/k: scroll  ctrl+q: exit")
	m.help = m.help.SetBindings(m.keys.PassthroughBindings())
	m.layout = m.layout.SetPassthrough(true)
	return m
}

func (m AppModel) navigationHints() string {
	if m.tooSmall {
		return " ?: help"
	}
	switch m.cursorSec {
	case ui.SectionSessions:
		return " n: new  d: delete  i: enter  enter: activate  q: quit"
	case ui.SectionApps:
		return " enter: launch  d: kill  ctrl+e/g/d: shortcuts  q: quit"
	case ui.SectionTasks:
		return " enter: run  p: run PTY  r: refresh  R: recursive  f: favorite  q: quit"
	default:
		return " enter: projects  n: new session  ?: help  q: quit"
	}
}

func (m AppModel) exitToNavigation() AppModel {
	m.mode = ModeNavigation
	m.status = m.status.SetMode("NORMAL")
	m.status = m.status.SetHints(m.navigationHints())
	m.help = m.help.SetBindings(m.keys.NavigationBindings())
	m.layout = m.layout.SetPassthrough(false)
	return m
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m.updateShutdown(msg)
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = msg.Width < minWidth || msg.Height-1 < minHeight
		m.status = m.status.SetWidth(msg.Width)
		m.status = m.status.SetHints(m.navigationHints())
		if m.tooSmall {
			m.status = m.status.SetSidebarHidden(false)
			return m, nil
		}
		m.sidebarCollapsed = msg.Width < collapseWidth
		statusHeight := 1
		m.layout = m.layout.SetSize(msg.Width, msg.Height-statusHeight, m.sidebarCollapsed)
		m.status = m.status.SetSidebarHidden(m.sidebarCollapsed)
		m.help = m.help.SetSize(msg.Width, msg.Height)
		m.labelModal.SetSize(msg.Width, msg.Height)
		m.projectModal.SetSize(msg.Width, msg.Height)
		return m, nil

	case terminal.TermErrorMsg:
		return m.handleTermError(msg)

	case TaskDoneMsg:
		if msg.Idx >= 0 && msg.Idx < len(m.tasks) {
			if msg.ExitCode == 0 {
				m.tasks[msg.Idx].Status = ui.TaskCompleted
			} else {
				m.tasks[msg.Idx].Status = ui.TaskFailed
			}
			if m.activeTask == msg.Idx {
				tv := m.buildTaskView(msg.Idx)
				m.layout = m.layout.SetTaskView(&tv)
			}
		}
		return m.syncSidebar(), nil

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

	case ui.ProjectConfirmMsg:
		return m.switchProject(msg.Dir)

	case ui.ProjectRemoveMsg:
		m.state.RemoveProject(msg.Dir)
		m.state.Save()
		return m, nil

	case tea.KeyMsg:
		if m.projectModal.Visible() {
			mod, modalCmd := m.projectModal.Update(msg)
			m.projectModal = *mod
			return m, modalCmd
		}

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

func (m AppModel) updateNavigationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "q":
		m, cmd := m.startShutdown()
		return m, cmd

	case msg.String() == "?":
		m.help = m.help.Toggle()
		return m, nil

	case msg.String() == "0":
		m.showInfo = true
		m.cursorSec = ui.SectionProjectInfo
		m.cursorIdx = 0
		m.focus = ui.FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		m.layout = m.layout.ShowInfo(true)
		return m.syncSidebar(), nil

	case key.Matches(msg, m.keys.FocusLeft):
		m.focus = ui.FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		return m, nil

	case key.Matches(msg, m.keys.FocusRight):
		m.focus = ui.FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		return m, nil

	case msg.String() == "tab":
		if m.focus == ui.FocusSidebar {
			m.focus = ui.FocusMain
			m.layout = m.layout.SetFocus(ui.FocusMain)
		} else {
			m.focus = ui.FocusSidebar
			m.layout = m.layout.SetFocus(ui.FocusSidebar)
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.focus == ui.FocusSidebar {
			var moveCmd tea.Cmd
			m, moveCmd = m.moveCursor(1)
			return m, moveCmd
		}
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.focus == ui.FocusSidebar {
			var moveCmd tea.Cmd
			m, moveCmd = m.moveCursor(-1)
			return m, moveCmd
		}
		return m, nil

	case msg.String() == "enter":
		if m.focus == ui.FocusSidebar {
			return m.activateCursor()
		}
		return m, nil

	case key.Matches(msg, m.keys.ToggleSidebar):
		m.sidebarCollapsed = !m.sidebarCollapsed
		statusHeight := 1
		m.layout = m.layout.SetSize(m.width, m.height-statusHeight, m.sidebarCollapsed)
		m.status = m.status.SetSidebarHidden(m.sidebarCollapsed)
		return m, nil

	case msg.String() == "ctrl+e":
		return m.activateAppByName("editor")

	case msg.String() == "ctrl+g":
		return m.activateAppByName("git")

	case msg.String() == "ctrl+d":
		return m.activateAppByName("docker")

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
		case ui.SectionTasks:
			return m.killTask(m.cursorIdx)
		}
		return m, nil

	case msg.String() == "i":
		sess := m.displayedSession()
		if sess == nil {
			return m, nil
		}
		m.focus = ui.FocusMain
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
		m.focus = ui.FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		m.cursorSec = ui.SectionSessions
		m.cursorIdx = 0
		return m.syncSidebar(), nil

	case msg.String() == "2":
		m.focus = ui.FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		m.cursorSec = ui.SectionApps
		m.cursorIdx = 0
		return m.syncSidebar(), nil

	case msg.String() == "3":
		m.focus = ui.FocusSidebar
		m.layout = m.layout.SetFocus(ui.FocusSidebar)
		m.cursorSec = ui.SectionTasks
		m.cursorIdx = 0
		return m.syncSidebar(), nil

	case msg.String() == "r":
		if m.cursorSec == ui.SectionTasks {
			m = m.refreshTasks(false)
			return m.syncSidebar(), nil
		}
		return m, nil

	case msg.String() == "R":
		if m.cursorSec == ui.SectionTasks {
			m = m.refreshTasks(true)
			return m.syncSidebar(), nil
		}
		return m, nil

	case msg.String() == "p":
		if m.cursorSec == ui.SectionTasks {
			return m.activateTask(m.cursorIdx, true)
		}
		return m, nil

	case msg.String() == "f":
		if m.cursorSec == ui.SectionTasks && m.cursorIdx >= 0 && m.cursorIdx < len(m.tasks) {
			taskKey := m.tasks[m.cursorIdx].Task.Key()
			m.state.TogglePreferred(m.projectDir, taskKey)
			m.state.Save()
			m = m.reorderTasks()
			return m.syncSidebar(), nil
		}
		return m, nil

	case key.Matches(msg, m.keys.PageUp):
		var scrollCmd tea.Cmd
		m.layout, scrollCmd = m.layout.ScrollMainPanel(-1)
		return m, scrollCmd
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

func (m AppModel) switchProject(dir string) (tea.Model, tea.Cmd) {
	for _, s := range m.sessions {
		m.layout = m.layout.RemoveSessionView(s.ID())
		s.Close()
	}
	m.sessions = nil
	m.activeIdx = -1

	for i := range m.apps {
		if m.apps[i].Sess != nil {
			m.layout = m.layout.RemoveSessionView(m.apps[i].Sess.ID())
			m.apps[i].Sess.Close()
			m.apps[i].Sess = nil
		}
	}
	m.activeApp = -1

	for i := range m.tasks {
		if m.tasks[i].Sess != nil {
			m.layout = m.layout.RemoveSessionView(m.tasks[i].Sess.ID())
			m.tasks[i].Sess.Close()
			m.tasks[i].Sess = nil
		}
		if m.tasks[i].cmd != nil && m.tasks[i].cmd.Process != nil {
			m.tasks[i].cmd.Process.Kill()
		}
	}
	m.activeTask = -1
	m.layout = m.layout.ClearTaskView()

	os.Chdir(dir)
	m.projectDir = dir
	m.projName, m.gitBranch = readProjectInfo(dir)
	m.state.RecordProject(dir)
	m.state.Save()

	if stored := m.state.LoadProjectTasks(dir); len(stored) > 0 {
		raw := task.TasksFromStored(stored)
		sorted := task.SortWithPreferred(raw, m.state.PreferredTasks[dir])
		m.tasks = make([]TaskRun, len(sorted))
		for i, t := range sorted {
			m.tasks[i] = TaskRun{Task: t, Status: ui.TaskPending}
		}
	} else {
		scanned := task.ScanTasks(dir, m.config.Tasks.PackageManager)
		m.state.SaveProjectTasks(dir, task.TasksToStored(scanned))
		sorted := task.SortWithPreferred(scanned, m.state.PreferredTasks[dir])
		m.tasks = make([]TaskRun, len(sorted))
		for i, t := range sorted {
			m.tasks[i] = TaskRun{Task: t, Status: ui.TaskPending}
		}
	}

	m.showInfo = true
	m.layout = m.layout.ShowInfo(true)
	m.layout = m.layout.SetProjectName(m.projName)
	m.layout = m.layout.SetKeyBindingGroups(DefaultKeyMap.ImportantBindingGroups())
	m.cursorSec = ui.SectionProjectInfo
	m.cursorIdx = 0
	m.focus = ui.FocusSidebar
	m.layout = m.layout.SetFocus(ui.FocusSidebar)

	return m.syncSidebar(), nil
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
		m.projName, m.gitBranch = readProjectInfo(m.projectDir)
		return m.syncSidebar(), nil
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	if m.tooSmall {
		return renderTooSmallWarning(m.width, m.height) + "\n" + m.status.View()
	}

	if m.width == 0 {
		return ""
	}

	if m.quitting {
		return m.shutdownView()
	}

	if m.projectModal.Visible() {
		return m.projectModal.View()
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

func renderTooSmallWarning(w, h int) string {
	msg := "Terminal too small — resize to at least 55×12"
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(ui.ColorText).Render(msg))
}
