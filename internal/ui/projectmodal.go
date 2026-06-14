package ui

import (
	"os"
	"path/filepath"
	"sort"
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

type dirEntry struct {
	name  string
	path  string
	isGit bool
}

type ProjectModal struct {
	visible   bool
	projects  []string
	cursorIdx int
	width     int
	height    int

	mode        modalMode
	browseDir   string
	entries     []dirEntry
	scrollOff   int
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

func (m *ProjectModal) enterBrowseMode() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	m.mode = modeBrowse
	m.browseDir = home
	m.cursorIdx = 0
	m.scrollOff = 0
	m.refreshEntries()
}

func (m *ProjectModal) refreshEntries() {
	tmp, err := readDirEntries(m.browseDir)
	if err != nil {
		m.entries = nil
		return
	}
	parent := filepath.Dir(m.browseDir)
	if parent != m.browseDir {
		tmp = append([]dirEntry{{name: "..", path: parent}}, tmp...)
	}
	m.entries = tmp
	if m.cursorIdx >= len(m.entries) {
		m.cursorIdx = 0
	}
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

func (m *ProjectModal) updateBrowse(msg tea.KeyMsg) (*ProjectModal, tea.Cmd) {
	switch {
	case msg.String() == "esc":
		m.mode = modeRecent
		m.cursorIdx = 0
		return m, nil
	case msg.String() == "enter" || msg.String() == " ":
		if m.cursorIdx < 0 || m.cursorIdx >= len(m.entries) {
			return m, nil
		}
		entry := m.entries[m.cursorIdx]
		if entry.name == ".." {
			m.browseDir = entry.path
			m.cursorIdx = 0
			m.scrollOff = 0
			m.refreshEntries()
			return m, nil
		}
		if entry.isGit {
			m.visible = false
			m.mode = modeRecent
			return m, func() tea.Msg {
				return ProjectConfirmMsg{Dir: entry.path}
			}
		}
		m.browseDir = entry.path
		m.cursorIdx = 0
		m.scrollOff = 0
		m.refreshEntries()
		return m, nil
	case msg.String() == "up" || msg.String() == "k":
		if m.cursorIdx > 0 {
			m.cursorIdx--
		}
	case msg.String() == "down" || msg.String() == "j":
		if m.cursorIdx < len(m.entries)-1 {
			m.cursorIdx++
		}
	case msg.String() == "backspace":
		parent := filepath.Dir(m.browseDir)
		if parent != m.browseDir {
			m.browseDir = parent
			m.cursorIdx = 0
			m.scrollOff = 0
			m.refreshEntries()
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
			MutedStyle.Render("Run lazycode from a git repository"),
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
		Width(44).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}

func (m *ProjectModal) viewBrowse() string {
	title := TitleStyle.Render("Switch Project")

	dirLine := MutedStyle.Render(m.browseDir)

	maxVisible := m.height - 10
	if maxVisible < 3 {
		maxVisible = 3
	}

	if m.cursorIdx < m.scrollOff {
		m.scrollOff = m.cursorIdx
	}
	if m.cursorIdx >= m.scrollOff+maxVisible {
		m.scrollOff = m.cursorIdx - maxVisible + 1
	}

	var itemLines []string
	end := m.scrollOff + maxVisible
	if end > len(m.entries) {
		end = len(m.entries)
	}
	for _, entry := range m.entries[m.scrollOff:end] {
		idx := m.scrollOff + len(itemLines)
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(ColorText)
		if idx == m.cursorIdx {
			prefix = "▸ "
			style = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
		}
		var label string
		if entry.isGit {
			label = style.Render(prefix + "\u25CF " + entry.name + "/")
		} else {
			label = style.Render(prefix + entry.name + "/")
		}
		itemLines = append(itemLines, label)
	}

	if len(m.entries) > end {
		itemLines = append(itemLines, MutedStyle.Render("  "+strings.Repeat("\u2500", 36)))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		dirLine,
		MutedStyle.Render(strings.Repeat("─", 38)),
		strings.Join(itemLines, "\n"),
		"",
		MutedStyle.Render("↑↓ navigate  Enter select  Esc back"),
	)

	box := BorderFocused.
		Width(44).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}

func readDirEntries(dir string) ([]dirEntry, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []dirEntry
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		_, gitErr := os.Stat(filepath.Join(full, ".git"))
		isGit := gitErr == nil
		result = append(result, dirEntry{
			name:  e.Name(),
			path:  full,
			isGit: isGit,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].isGit != result[j].isGit {
			return result[i].isGit
		}
		return result[i].name < result[j].name
	})

	return result, nil
}
