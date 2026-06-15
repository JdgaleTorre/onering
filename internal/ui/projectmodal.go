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

type ProjectRemoveMsg struct {
	Dir string
}

type modalMode int

const (
	modeRecent modalMode = iota
	modeBrowse
)

type ProjectModal struct {
	visible   bool
	projects  []string
	cursorIdx int
	width     int
	height    int

	mode      modalMode
	browseDir string
	entries   []dirEntry
	scrollOff int
}

func NewProjectModal() ProjectModal {
	return ProjectModal{}
}

func (m *ProjectModal) Open(projects []string) {
	m.visible = true
	m.projects = projects
	m.cursorIdx = 0
	m.mode = modeRecent
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
		switch m.mode {
		case modeRecent:
			return m.updateRecent(msg)
		case modeBrowse:
			return m.updateBrowse(msg)
		}
	}

	return m, nil
}

func (m *ProjectModal) updateRecent(msg tea.KeyMsg) (*ProjectModal, tea.Cmd) {
	switch {
	case msg.String() == "esc":
		m.visible = false
		return m, nil
	case msg.String() == "enter" || msg.String() == " ":
		if m.cursorIdx == len(m.projects) {
			m.enterBrowseMode()
			return m, nil
		}
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
		total := len(m.projects) + 1
		if m.cursorIdx < total-1 {
			m.cursorIdx++
		}
	case msg.String() == "d":
		if m.cursorIdx < len(m.projects) {
			dir := m.projects[m.cursorIdx]
			m.projects = append(m.projects[:m.cursorIdx], m.projects[m.cursorIdx+1:]...)
			if m.cursorIdx >= len(m.projects) && m.cursorIdx > 0 {
				m.cursorIdx--
			}
			return m, func() tea.Msg {
				return ProjectRemoveMsg{Dir: dir}
			}
		}
	}
	return m, nil
}

func (m *ProjectModal) View() string {
	if !m.visible {
		return ""
	}
	switch m.mode {
	case modeRecent:
		return m.viewRecent()
	case modeBrowse:
		return m.viewBrowse()
	}
	return ""
}

func (m *ProjectModal) viewRecent() string {
	title := TitleStyle.Render("Switch Project")

	var items []string
	if len(m.projects) == 0 {
		items = append(items,
			MutedStyle.Render("No recent projects."),
			"",
			MutedStyle.Render("Run onering from a git repository"),
			MutedStyle.Render("to get started."))
	} else {
		for i, proj := range m.projects {
			name := filepath.Base(proj)
			lineStyle := lipgloss.NewStyle().Foreground(ColorText)
			prefix := "  "
			if i == m.cursorIdx {
				prefix = "▸ "
				lineStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
			}
			items = append(items, lineStyle.Render(prefix+name))
		}
	}

	browseStyle := lipgloss.NewStyle().Foreground(ColorText)
	browsePrefix := "  "
	if m.cursorIdx == len(m.projects) {
		browsePrefix = "▸ "
		browseStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	}
	items = append(items,
		"",
		MutedStyle.Render(strings.Repeat("─", 38)),
		browseStyle.Render(browsePrefix+"Browse filesystem…"))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strings.Join(items, "\n"),
		"",
		MutedStyle.Render("↑↓ navigate  Enter select  d delete from recents  Esc cancel"),
	)

	box := BorderFocused.
		Width(ModalWidth).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}
