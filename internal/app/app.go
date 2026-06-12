package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/josegale/lazycode/internal/config"
	"github.com/josegale/lazycode/internal/ui"
)

type AppModel struct {
	config *config.Config
	keys   KeyMap
	mode   InputMode
	focus  FocusPanel
	layout ui.LayoutModel
	status ui.StatusBarModel
	help   ui.HelpModel
	width  int
	height int
}

func New(cfg *config.Config) AppModel {
	return AppModel{
		config: cfg,
		keys:   DefaultKeyMap,
		mode:   ModeNavigation,
		focus:  FocusSidebar,
		layout: ui.NewLayoutModel(cfg),
		status: ui.NewStatusBarModel(),
		help:   ui.NewHelpModel(DefaultKeyMap.NavigationBindings()),
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		statusHeight := 1
		m.layout = m.layout.SetSize(msg.Width, msg.Height-statusHeight)
		m.status = m.status.SetWidth(msg.Width)
		m.help = m.help.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.help.Visible() {
			var cmd tea.Cmd
			m.help, cmd = m.help.Update(msg)
			return m, cmd
		}

		if m.mode == ModeInsert {
			return m.updateInsertMode(msg)
		}
		return m.updateNavigationMode(msg)
	}

	return m, nil
}

func (m AppModel) updateNavigationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "?":
		m.help = m.help.Toggle()
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
	case msg.String() == "i":
		m.mode = ModeInsert
		m.focus = FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		m.status = m.status.SetMode("INSERT")
		m.help = m.help.SetBindings(m.keys.InsertBindings())
		return m, nil
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

func (m AppModel) updateInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = ModeNavigation
		m.status = m.status.SetMode("NORMAL")
		m.help = m.help.SetBindings(m.keys.NavigationBindings())
		return m, nil
	}

	var cmd tea.Cmd
	m.layout, cmd = m.layout.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	if m.width == 0 {
		return ""
	}

	view := m.layout.View() + "\n" + m.status.View()

	if m.help.Visible() {
		return m.help.View()
	}

	return view
}
