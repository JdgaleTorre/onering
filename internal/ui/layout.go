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
	FocusAgent
)

type LayoutMode int

const (
	LayoutSingle LayoutMode = iota
	LayoutSplit
)

const (
	ResizeSmallStep    = 5
	ResizeLargeStep    = 15
	MinSidebarWidth    = 20
	MinPanelWidth      = 20
	MaxSidebarFraction = 0.6
)

type LayoutModel struct {
	sidebar          SidebarModel
	mainPanel        MainPanelModel
	agentPanel       AgentPanelModel
	focus            FocusPanel
	layoutMode       LayoutMode
	width            int
	height           int
	sidebarW         int
	agentFraction    float64
	sidebarCollapsed bool
}

func NewLayoutModel(cfg *config.Config) LayoutModel {
	return LayoutModel{
		sidebar:       NewSidebarModel(),
		mainPanel:     NewMainPanelModel(),
		agentPanel:    NewAgentPanelModel(),
		focus:         FocusSidebar,
		layoutMode:    LayoutSingle,
		sidebarW:      cfg.UI.SidebarWidth,
		agentFraction: 1.0 / 3.0,
	}
}

func (m LayoutModel) LayoutMode() LayoutMode {
	return m.layoutMode
}

func (m LayoutModel) SetLayoutMode(mode LayoutMode) LayoutModel {
	m.layoutMode = mode
	return m.SetSize(m.width, m.height, m.sidebarCollapsed)
}

func (m LayoutModel) SetSize(w, h int, sidebarCollapsed bool) LayoutModel {
	m.width = w
	m.height = h
	m.sidebarCollapsed = sidebarCollapsed

	if sidebarCollapsed {
		m.sidebar = m.sidebar.SetSize(0, h)
		if m.layoutMode == LayoutSplit {
			mainW, agentW := m.splitRemaining(w - 1)
			m.mainPanel = m.mainPanel.SetSize(mainW, h)
			m.agentPanel = m.agentPanel.SetSize(agentW, h)
		} else {
			m.mainPanel = m.mainPanel.SetSize(w, h)
		}
	} else {
		sideW := m.sidebarW
		if m.layoutMode == LayoutSplit {
			maxSide := w - 2*MinPanelWidth - 2
			if maxSide < MinSidebarWidth {
				maxSide = MinSidebarWidth
			}
			if sideW > maxSide {
				sideW = maxSide
			}
			remaining := w - sideW - 1
			mainW, agentW := m.splitRemaining(remaining - 1)
			m.sidebar = m.sidebar.SetSize(sideW, h)
			m.mainPanel = m.mainPanel.SetSize(mainW, h)
			m.agentPanel = m.agentPanel.SetSize(agentW, h)
		} else {
			maxW := w - MinPanelWidth - 1
			if maxW < MinSidebarWidth {
				maxW = MinSidebarWidth
			}
			if sideW > maxW {
				sideW = maxW
			}
			mainW := w - sideW - 1
			if mainW < 1 {
				mainW = 1
			}
			m.sidebar = m.sidebar.SetSize(sideW, h)
			m.mainPanel = m.mainPanel.SetSize(mainW, h)
		}
	}
	return m
}

func (m LayoutModel) splitRemaining(remaining int) (int, int) {
	agentW := int(m.agentFraction * float64(remaining))
	if agentW < MinPanelWidth {
		agentW = MinPanelWidth
	}
	mainW := remaining - agentW
	if mainW < MinPanelWidth {
		mainW = MinPanelWidth
		agentW = remaining - mainW
		if agentW < 1 {
			agentW = 1
		}
	}
	return mainW, agentW
}

func (m LayoutModel) ResizeSidebar(delta int) LayoutModel {
	newW := m.sidebarW + delta

	if newW < MinSidebarWidth {
		newW = MinSidebarWidth
	}

	var maxW int
	if m.layoutMode == LayoutSplit {
		maxW = m.width - 2*MinPanelWidth - 2
	} else {
		maxByFraction := int(MaxSidebarFraction * float64(m.width))
		maxByMainMin := m.width - MinPanelWidth - 1
		maxW = maxByFraction
		if maxByMainMin < maxW {
			maxW = maxByMainMin
		}
	}
	if maxW < MinSidebarWidth {
		maxW = MinSidebarWidth
	}
	if newW > maxW {
		newW = maxW
	}

	m.sidebarW = newW
	return m.SetSize(m.width, m.height, m.sidebarCollapsed)
}

func (m LayoutModel) ResizeAgentPanel(delta int) LayoutModel {
	if m.layoutMode != LayoutSplit {
		return m
	}

	var remaining int
	if m.sidebarCollapsed {
		remaining = m.width - 1
	} else {
		remaining = m.width - m.sidebarW - 2
	}

	currentAgentW := int(m.agentFraction * float64(remaining))
	newAgentW := currentAgentW + delta
	if newAgentW < MinPanelWidth {
		newAgentW = MinPanelWidth
	}
	newMainW := remaining - newAgentW
	if newMainW < MinPanelWidth {
		newMainW = MinPanelWidth
		newAgentW = remaining - newMainW
	}
	if remaining > 0 {
		m.agentFraction = float64(newAgentW) / float64(remaining)
	}
	return m.SetSize(m.width, m.height, m.sidebarCollapsed)
}

func (m LayoutModel) SetSidebarCollapsed(collapsed bool) LayoutModel {
	m.sidebarCollapsed = collapsed
	return m
}

func (m LayoutModel) SetFocus(f FocusPanel) LayoutModel {
	m.focus = f
	m.sidebar = m.sidebar.SetFocused(f == FocusSidebar)
	m.mainPanel = m.mainPanel.SetFocused(f == FocusMain)
	m.agentPanel = m.agentPanel.SetFocused(f == FocusAgent)
	return m
}

func (m LayoutModel) SetSidebar(d SidebarData) LayoutModel {
	m.sidebar = m.sidebar.SetData(d)
	return m
}

func (m LayoutModel) SetPassthrough(b bool) LayoutModel {
	m.mainPanel = m.mainPanel.SetPassthrough(b)
	m.agentPanel = m.agentPanel.SetPassthrough(b)
	return m
}

func (m LayoutModel) SetActiveSession(sess agent.Session) (LayoutModel, tea.Cmd) {
	mp, cmd := m.mainPanel.SetSession(sess)
	m.mainPanel = mp
	return m, cmd
}

func (m LayoutModel) SetAgentSession(sess agent.Session) (LayoutModel, tea.Cmd) {
	ap, cmd := m.agentPanel.SetSession(sess)
	m.agentPanel = ap
	return m, cmd
}

func (m LayoutModel) MainPanel() MainPanelModel {
	return m.mainPanel
}

func (m LayoutModel) AgentPanel() AgentPanelModel {
	return m.agentPanel
}

func (m LayoutModel) RemoveSessionView(sessionID string) LayoutModel {
	m.mainPanel = m.mainPanel.RemoveSessionView(sessionID)
	return m
}

func (m LayoutModel) RemoveAgentSessionView(sessionID string) LayoutModel {
	m.agentPanel = m.agentPanel.RemoveSessionView(sessionID)
	return m
}

func (m LayoutModel) TransferToAgent(sessionID string) LayoutModel {
	mp, tv, ok := m.mainPanel.DetachTermView(sessionID)
	if !ok {
		return m
	}
	m.mainPanel = mp
	m.agentPanel = m.agentPanel.AttachTermView(sessionID, tv)
	return m
}

func (m LayoutModel) TransferToMain(sessionID string) LayoutModel {
	ap, tv, ok := m.agentPanel.DetachTermView(sessionID)
	if !ok {
		return m
	}
	m.agentPanel = ap
	m.mainPanel = m.mainPanel.AttachTermView(sessionID, tv)
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

func (m LayoutModel) ScrollAgentPanel(direction int) (LayoutModel, tea.Cmd) {
	ap, cmd := m.agentPanel.ScrollTermView(direction)
	m.agentPanel = ap
	return m, cmd
}

func (m LayoutModel) Update(msg tea.Msg) (LayoutModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.(type) {
	case terminal.OutputMsg, terminal.TermErrorMsg, terminal.ClearToastMsg, terminal.ColorSchemeChangedMsg:
		var cmd1, cmd2 tea.Cmd
		m.mainPanel, cmd1 = m.mainPanel.Update(msg)
		if m.layoutMode == LayoutSplit {
			m.agentPanel, cmd2 = m.agentPanel.Update(msg)
		}
		return m, tea.Batch(cmd1, cmd2)
	}

	switch msg := msg.(type) {
	case tea.MouseMsg:
		return m.handleMouse(msg)
	default:
		switch m.focus {
		case FocusSidebar:
			m.sidebar, cmd = m.sidebar.Update(msg)
		case FocusMain:
			m.mainPanel, cmd = m.mainPanel.Update(msg)
		case FocusAgent:
			if m.layoutMode == LayoutSplit {
				m.agentPanel, cmd = m.agentPanel.Update(msg)
			}
		}
	}

	return m, cmd
}

func (m LayoutModel) handleMouse(msg tea.MouseMsg) (LayoutModel, tea.Cmd) {
	var cmd tea.Cmd
	sideW := 0
	if !m.sidebarCollapsed {
		sideW = m.sidebarW + 1
	}

	if m.layoutMode == LayoutSplit {
		var remaining int
		if m.sidebarCollapsed {
			remaining = m.width - 1
		} else {
			remaining = m.width - sideW - 1
		}
		mainW, _ := m.splitRemaining(remaining)
		mainStart := sideW
		mainEnd := mainStart + mainW
		agentStart := mainEnd + 1

		contentH := m.height - 2
		adjY := msg.Y - 1

		if msg.X >= mainStart && msg.X < mainEnd {
			adjX := msg.X - mainStart - 1
			contentW := mainW - 2
			if adjX >= 0 && adjX < contentW && adjY >= 0 && adjY < contentH {
				msg.X = adjX
				msg.Y = adjY
				m.mainPanel, cmd = m.mainPanel.Update(msg)
			}
		} else if msg.X >= agentStart {
			adjX := msg.X - agentStart - 1
			agentW := m.width - agentStart
			contentW := agentW - 2
			if adjX >= 0 && adjX < contentW && adjY >= 0 && adjY < contentH {
				msg.X = adjX
				msg.Y = adjY
				m.agentPanel, cmd = m.agentPanel.Update(msg)
			}
		}
	} else {
		mainW := m.width - sideW
		adjX := msg.X - sideW - 1
		contentW := mainW - 2
		contentH := m.height - 2
		adjY := msg.Y - 1
		if adjX >= 0 && adjX < contentW && adjY >= 0 && adjY < contentH {
			msg.X = adjX
			msg.Y = adjY
			m.mainPanel, cmd = m.mainPanel.Update(msg)
		}
	}

	return m, cmd
}

func (m LayoutModel) View() string {
	if m.sidebarCollapsed {
		if m.layoutMode == LayoutSplit {
			return lipgloss.JoinHorizontal(
				lipgloss.Top,
				m.mainPanel.View(),
				" ",
				m.agentPanel.View(),
			)
		}
		return m.mainPanel.View()
	}
	if m.layoutMode == LayoutSplit {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.sidebar.View(),
			" ",
			m.mainPanel.View(),
			" ",
			m.agentPanel.View(),
		)
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.sidebar.View(),
		" ",
		m.mainPanel.View(),
	)
}
