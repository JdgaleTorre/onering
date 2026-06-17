package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/JdgaleTorre/onering/internal/agent"
	"github.com/JdgaleTorre/onering/internal/config"
	"github.com/JdgaleTorre/onering/internal/terminal"
)

type FocusPanel int

const (
	FocusSidebar FocusPanel = iota
	FocusMain
)

type LayoutModel struct {
	sidebar          SidebarModel
	mainPanel        MainPanelModel
	focus            FocusPanel
	width            int
	height           int
	sidebarW         int
	sidebarCollapsed bool
}

func NewLayoutModel(cfg *config.Config) LayoutModel {
	return LayoutModel{
		sidebar:   NewSidebarModel(),
		mainPanel: NewMainPanelModel(),
		focus:     FocusSidebar,
		sidebarW:  cfg.UI.SidebarWidth,
	}
}

func (m LayoutModel) SetSize(w, h int, sidebarCollapsed bool) LayoutModel {
	m.width = w
	m.height = h
	m.sidebarCollapsed = sidebarCollapsed

	if sidebarCollapsed {
		m.sidebar = m.sidebar.SetSize(0, h)
		m.mainPanel = m.mainPanel.SetSize(w, h)
	} else {
		sideW := m.sidebarW
		mainW := w - sideW - 1
		if mainW < 1 {
			mainW = 1
		}
		m.sidebar = m.sidebar.SetSize(sideW, h)
		m.mainPanel = m.mainPanel.SetSize(mainW, h)
	}
	return m
}

func (m LayoutModel) SetSidebarCollapsed(collapsed bool) LayoutModel {
	m.sidebarCollapsed = collapsed
	return m
}

func (m LayoutModel) SetFocus(f FocusPanel) LayoutModel {
	m.focus = f
	m.sidebar = m.sidebar.SetFocused(f == FocusSidebar)
	m.mainPanel = m.mainPanel.SetFocused(f == FocusMain)
	return m
}

func (m LayoutModel) SetSidebar(d SidebarData) LayoutModel {
	m.sidebar = m.sidebar.SetData(d)
	return m
}

func (m LayoutModel) SetPassthrough(b bool) LayoutModel {
	m.mainPanel = m.mainPanel.SetPassthrough(b)
	return m
}

func (m LayoutModel) SetActiveSession(sess agent.Session) (LayoutModel, tea.Cmd) {
	mp, cmd := m.mainPanel.SetSession(sess)
	m.mainPanel = mp
	return m, cmd
}

func (m LayoutModel) MainPanel() MainPanelModel {
	return m.mainPanel
}

func (m LayoutModel) RemoveSessionView(sessionID string) LayoutModel {
	m.mainPanel = m.mainPanel.RemoveSessionView(sessionID)
	return m
}

func (m LayoutModel) ShowInfo(show bool) LayoutModel {
	m.mainPanel = m.mainPanel.ShowInfo(show)
	return m
}

func (m LayoutModel) SetInfoText(text string) LayoutModel {
	m.mainPanel = m.mainPanel.SetInfoText(text)
	return m
}

func (m LayoutModel) ClearInfoText() LayoutModel {
	m.mainPanel = m.mainPanel.ClearInfoText()
	return m
}

func (m LayoutModel) SetKeyBindingGroups(groups []BindingGroup) LayoutModel {
	m.mainPanel = m.mainPanel.SetKeyBindingGroups(groups)
	return m
}

func (m LayoutModel) SetProjectName(name string) LayoutModel {
	m.mainPanel = m.mainPanel.SetProjectName(name)
	return m
}

func (m LayoutModel) MainPanelTermSize() (int, int) {
	return m.mainPanel.termSize()
}

func (m LayoutModel) SetTaskView(tv *TaskViewModel) LayoutModel {
	m.mainPanel = m.mainPanel.SetTaskView(tv)
	return m
}

func (m LayoutModel) ClearTaskView() LayoutModel {
	m.mainPanel = m.mainPanel.ClearTaskView()
	return m
}

func (m LayoutModel) ScrollMainPanel(direction int) (LayoutModel, tea.Cmd) {
	mp, cmd := m.mainPanel.ScrollTermView(direction)
	m.mainPanel = mp
	return m, cmd
}

func (m LayoutModel) Update(msg tea.Msg) (LayoutModel, tea.Cmd) {
	var cmd tea.Cmd

	// Terminal output must always reach the main panel; routing it by
	// focus would drop it and kill the session's read loop.
	switch msg.(type) {
	case terminal.OutputMsg, terminal.TermErrorMsg, terminal.ClearToastMsg, terminal.ColorSchemeChangedMsg:
		m.mainPanel, cmd = m.mainPanel.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.MouseMsg:
		var mainW int
		var adjX int
		if m.sidebarCollapsed {
			mainW = m.width
			adjX = msg.X - 1
		} else {
			mainW = m.width - m.sidebarW - 1
			adjX = msg.X - m.sidebarW - 2
		}
		contentW := mainW - 2
		contentH := m.height - 2
		adjY := msg.Y - 1
		if adjX < 0 || adjX >= contentW || adjY < 0 || adjY >= contentH {
			return m, nil
		}
		msg.X = adjX
		msg.Y = adjY
		m.mainPanel, cmd = m.mainPanel.Update(msg)
	default:
		switch m.focus {
		case FocusSidebar:
			m.sidebar, cmd = m.sidebar.Update(msg)
		case FocusMain:
			m.mainPanel, cmd = m.mainPanel.Update(msg)
		}
	}

	return m, cmd
}

func (m LayoutModel) View() string {
	if m.sidebarCollapsed {
		return m.mainPanel.View()
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.sidebar.View(),
		" ",
		m.mainPanel.View(),
	)
}
