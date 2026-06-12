package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/josegale/lazycode/internal/config"
)

type FocusPanel int

const (
	FocusSidebar FocusPanel = iota
	FocusMain
)

type LayoutModel struct {
	sidebar   SidebarModel
	mainPanel MainPanelModel
	focus     FocusPanel
	width     int
	height    int
	sidebarW  int
}

func NewLayoutModel(cfg *config.Config) LayoutModel {
	return LayoutModel{
		sidebar:   NewSidebarModel(),
		mainPanel: NewMainPanelModel(),
		focus:     FocusSidebar,
		sidebarW:  cfg.UI.SidebarWidth,
	}
}

func (m LayoutModel) SetSize(w, h int) LayoutModel {
	m.width = w
	m.height = h

	sideW := m.sidebarW
	mainW := w - sideW - 1 // 1 for the gap

	m.sidebar = m.sidebar.SetSize(sideW, h)
	m.mainPanel = m.mainPanel.SetSize(mainW, h)
	return m
}

func (m LayoutModel) SetFocus(f FocusPanel) LayoutModel {
	m.focus = f
	m.sidebar = m.sidebar.SetFocused(f == FocusSidebar)
	m.mainPanel = m.mainPanel.SetFocused(f == FocusMain)
	return m
}

func (m LayoutModel) Update(msg tea.Msg) (LayoutModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {
	case FocusSidebar:
		m.sidebar, cmd = m.sidebar.Update(msg)
	case FocusMain:
		m.mainPanel, cmd = m.mainPanel.Update(msg)
	}

	return m, cmd
}

func (m LayoutModel) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.sidebar.View(),
		" ",
		m.mainPanel.View(),
	)
}
