package ui

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProjectConfirmMsg struct {
	Dir string
}

type ProjectModal struct {
	visible   bool
	projects  []string
	cursorIdx int
	width     int
	height    int
}

func NewProjectModal() ProjectModal {
	return ProjectModal{}
}

func (m *ProjectModal) Open(projects []string) {
	m.visible = true
	m.projects = projects
	m.cursorIdx = 0
}

func (m *ProjectModal) Close() {
	m.visible = false
}

func (m *ProjectModal) Visible() bool {
	return m.visible
}

func (m *ProjectModal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *ProjectModal) Update(msg tea.Msg) (*ProjectModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "esc":
			m.visible = false
			return m, nil
		case msg.String() == "enter" || msg.String() == " ":
			if m.cursorIdx >= 0 && m.cursorIdx < len(m.projects) {
				dir := m.projects[m.cursorIdx]
				m.visible = false
				return m, func() tea.Msg {
					return ProjectConfirmMsg{Dir: dir}
				}
			}
		case msg.String() == "up" || msg.String() == "k":
			if m.cursorIdx > 0 {
				m.cursorIdx--
			}
		case msg.String() == "down" || msg.String() == "j":
			if m.cursorIdx < len(m.projects)-1 {
				m.cursorIdx++
			}
		}
	}

	return m, nil
}

func (m *ProjectModal) View() string {
	if !m.visible {
		return ""
	}

	title := TitleStyle.Render("Switch Project")

	var lines []string
	if len(m.projects) == 0 {
		lines = append(lines, MutedStyle.Render("No recent projects."))
	} else {
		for i, proj := range m.projects {
			name := filepath.Base(proj)
			lineStyle := lipgloss.NewStyle().Foreground(ColorText)
			prefix := "  "
			if i == m.cursorIdx {
				prefix = "▸ "
				lineStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
			}
			lines = append(lines, lineStyle.Render(prefix+name))
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strings.Join(lines, "\n"),
		"",
		MutedStyle.Render("↑↓ navigate  Enter switch  Esc cancel"),
	)

	box := BorderFocused.
		Width(44).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}
