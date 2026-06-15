package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/josegale/onering/internal/agent"
)

type SidebarSection int

const (
	SectionProjectInfo SidebarSection = iota
	SectionSessions
	SectionApps
	SectionTasks
)

type AppItem struct {
	Name      string
	Running   bool
	Installed bool
}

type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
)

type TaskItem struct {
	Name      string
	Source    string
	Dir       string
	Status    TaskStatus
	Preferred bool
}

type SidebarData struct {
	ProjectName string
	Branch      string
	Sessions    []agent.Session
	Apps        []AppItem
	Tasks       []TaskItem
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
		Width(m.width-2).
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
	projectBox := m.renderSectionBox(0, "Project Info", projLines, m.data.CursorSection == SectionProjectInfo)

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
	const maxVisApps = 5
	n := len(m.data.Apps)
	appLines := make([]string, maxVisApps)

	start := 0
	if n > maxVisApps {
		start = m.data.CursorIdx - maxVisApps/2
		if start < 0 {
			start = 0
		}
		if start+maxVisApps > n {
			start = n - maxVisApps
		}
	}

	for i := 0; i < maxVisApps; i++ {
		idx := start + i
		if idx >= n {
			appLines[i] = ""
			continue
		}
		app := m.data.Apps[idx]
		cursorHere := m.data.CursorSection == SectionApps && m.data.CursorIdx == idx

		if !app.Installed {
			prefix := "  "
			if cursorHere {
				prefix = "▸ "
			}
			appLines[i] = prefix + MutedStyle.Render("!"+app.Name)
			continue
		}
		lineStyle := lipgloss.NewStyle().Foreground(ColorText)
		if cursorHere {
			lineStyle = lineStyle.Foreground(ColorPrimary).Bold(true)
		}
		badge := ""
		if app.Running {
			badge = RunningStyle.Render(" ●")
		}
		appLines[i] = m.cursorPrefix(SectionApps, idx) + lineStyle.Render(app.Name) + badge
	}
	appsBox := m.renderSectionBox(2, "Apps", appLines, m.data.CursorSection == SectionApps)

	// Tasks section
	var tasksBox string
	if tn := len(m.data.Tasks); tn > 0 {
		const maxVisTasks = 5
		taskLines := make([]string, maxVisTasks)
		tStart := 0
		if tn > maxVisTasks {
			tStart = m.data.CursorIdx - maxVisTasks/2
			if tStart < 0 {
				tStart = 0
			}
			if tStart+maxVisTasks > tn {
				tStart = tn - maxVisTasks
			}
		}
		for i := 0; i < maxVisTasks; i++ {
			tidx := tStart + i
			if tidx >= tn {
				taskLines[i] = ""
				continue
			}
			t := m.data.Tasks[tidx]
			cursorHere := m.data.CursorSection == SectionTasks && m.data.CursorIdx == tidx
			lineStyle := lipgloss.NewStyle().Foreground(ColorText)
			if cursorHere {
				lineStyle = lineStyle.Foreground(ColorPrimary).Bold(true)
			}
			badge := ""
			switch t.Status {
			case TaskRunning:
				badge = RunningStyle.Render(" ●")
			case TaskCompleted:
				badge = lipgloss.NewStyle().Foreground(ColorInsert).Render(" ✓")
			case TaskFailed:
				badge = lipgloss.NewStyle().Foreground(ColorError).Render(" ✗")
			}
			displayDir := t.Dir
			if !cursorHere && t.Dir != "" {
				if idx := strings.IndexByte(t.Dir, '/'); idx > 0 {
					displayDir = t.Dir[:idx]
				}
			}
			var prefix string
			if displayDir != "" {
				if t.Preferred {
					prefix = lipgloss.NewStyle().Foreground(ColorPrimary).Render("★ ") + MutedStyle.Render(displayDir+"/") + MutedStyle.Render(t.Source+": ")
				} else {
					prefix = MutedStyle.Render(displayDir+"/") + MutedStyle.Render(t.Source+": ")
				}
			} else {
				if t.Preferred {
					prefix = lipgloss.NewStyle().Foreground(ColorPrimary).Render("★ ") + MutedStyle.Render(t.Source+": ")
				} else {
					prefix = MutedStyle.Render(t.Source + ": ")
				}
			}
			taskLines[i] = m.cursorPrefix(SectionTasks, tidx) + prefix + lineStyle.Render(t.Name) + badge
			taskLines[i] = ansi.Truncate(taskLines[i], m.width-4, "…")
		}
		tasksBox = m.renderSectionBox(3, "Tasks", taskLines, m.data.CursorSection == SectionTasks)
	} else {
		taskLines := []string{MutedStyle.Render("No tasks found.")}
		tasksBox = m.renderSectionBox(3, "Tasks", taskLines, m.data.CursorSection == SectionTasks)
	}

	// Join all boxes vertically
	result := lipgloss.JoinVertical(lipgloss.Left, projectBox, sessionsBox, appsBox, tasksBox)

	// Clip to height
	resultLines := strings.Split(result, "\n")
	if len(resultLines) > m.height {
		resultLines = resultLines[:m.height]
	}

	return strings.Join(resultLines, "\n")
}
