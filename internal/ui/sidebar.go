package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/josegale/lazycode/internal/agent"
	"github.com/charmbracelet/x/ansi"
)

type SidebarSection int

const (
	SectionSessions SidebarSection = iota
	SectionApps
)

type AppItem struct {
	Name    string
	Running bool
}

type SidebarData struct {
	ProjectName string
	Branch      string
	Sessions []agent.Session
	Apps     []AppItem
	// Cursor position: which section it is in and the item index within it.
	CursorSection SidebarSection
	CursorIdx     int
}

type SidebarModel struct {
	focused bool
	width   int
	height  int
	data    SidebarData
}

func NewSidebarModel() SidebarModel {
	return SidebarModel{}
}

func (m SidebarModel) SetSize(w, h int) SidebarModel {
	m.width = w
	m.height = h
	return m
}

func (m SidebarModel) SetFocused(f bool) SidebarModel {
	m.focused = f
	return m
}

func (m SidebarModel) SetData(d SidebarData) SidebarModel {
	m.data = d
	return m
}

func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	return m, nil
}

func (m SidebarModel) cursorPrefix(sec SidebarSection, idx int) string {
	if m.data.CursorSection == sec && m.data.CursorIdx == idx {
		return "▸ "
	}
	return "  "
}

func (m SidebarModel) renderSectionBox(number int, title string, lines []string, active bool) string {
	borderColor := ColorBorder
	if active && m.focused {
		borderColor = ColorFocused
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Padding(0, 1)

	inner := strings.Join(lines, "\n")
	boxed := style.Render(inner)

	// Replace the top border with one that includes the label.
	boxLines := strings.Split(boxed, "\n")
	if len(boxLines) > 0 {
		// Calculate the label: "[N] Title"
		label := fmt.Sprintf("[%d] %s", number, title)
		// Get display width (ignoring ANSI codes)
		labelDisplayLen := ansi.StringWidth(label)

		// Available width for the top border:
		// The box is m.width wide. The top border is: ╭─label + dashes + ╮
		// Total width must equal m.width.
		// Format: ╭─[label]──────╮
		// Left part: "╭─" (2 chars)
		// Label: labelDisplayLen chars (display width, ignoring ANSI)
		// Right part: "╮" (1 char)
		// Middle dashes: m.width - 2 - 1 - labelDisplayLen
		dashCount := m.width - 2 - 1 - labelDisplayLen
		if dashCount < 0 {
			dashCount = 0
		}

		rawTop := "╭─" + label + strings.Repeat("─", dashCount) + "╮"
		boxLines[0] = lipgloss.NewStyle().Foreground(borderColor).Render(rawTop)
		boxed = strings.Join(boxLines, "\n")
	}

	return boxed
}

func (m SidebarModel) View() string {
	// Project Info section (no number, always inactive border)
	var projLines []string
	projLines = append(projLines, MutedStyle.Render(m.data.ProjectName))
	if m.data.Branch != "" {
		projLines = append(projLines, MutedStyle.Render("⎇ "+m.data.Branch))
	}
	projectBox := m.renderSectionBox(0, "Project Info", projLines, false)

	// Sessions section
	var sessionLines []string
	if len(m.data.Sessions) == 0 {
		sessionLines = append(sessionLines, MutedStyle.Render("No sessions yet."))
		sessionLines = append(sessionLines, MutedStyle.Render("Press 'n' to create one."))
	} else {
		for i, sess := range m.data.Sessions {
			label := fmt.Sprintf("%s: %s", sess.AgentName(), sess.Label())
			lineStyle := lipgloss.NewStyle().Foreground(ColorText)
			if m.data.CursorSection == SectionSessions && i == m.data.CursorIdx {
				lineStyle = lineStyle.Foreground(ColorPrimary).Bold(true)
			}
			badge := ""
			if _, isPTY := sess.(agent.PTYProvider); isPTY {
				badge = RunningStyle.Render(" PTY ")
			}
			sessionLines = append(sessionLines, m.cursorPrefix(SectionSessions, i)+lineStyle.Render(label)+badge)
		}
	}
	sessionsBox := m.renderSectionBox(1, "Sessions", sessionLines, m.data.CursorSection == SectionSessions)

	// Apps section
	var appLines []string
	if len(m.data.Apps) == 0 {
		appLines = append(appLines, MutedStyle.Render("None configured."))
	} else {
		for i, app := range m.data.Apps {
			lineStyle := lipgloss.NewStyle().Foreground(ColorText)
			if m.data.CursorSection == SectionApps && i == m.data.CursorIdx {
				lineStyle = lineStyle.Foreground(ColorPrimary).Bold(true)
			}
			badge := ""
			if app.Running {
				badge = RunningStyle.Render(" ●")
			}
			appLines = append(appLines, m.cursorPrefix(SectionApps, i)+lineStyle.Render(app.Name)+badge)
		}
	}
	appsBox := m.renderSectionBox(2, "Apps", appLines, m.data.CursorSection == SectionApps)

	// Join all boxes vertically
	result := lipgloss.JoinVertical(lipgloss.Left, projectBox, sessionsBox, appsBox)

	// Clip to height
	resultLines := strings.Split(result, "\n")
	if len(resultLines) > m.height {
		resultLines = resultLines[:m.height]
	}

	return strings.Join(resultLines, "\n")
}
