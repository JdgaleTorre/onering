package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HelpModel struct {
	bindings []key.Binding
	visible  bool
	width    int
	height   int
}

func NewHelpModel(bindings []key.Binding) HelpModel {
	return HelpModel{bindings: bindings}
}

func (m HelpModel) SetSize(w, h int) HelpModel {
	m.width = w
	m.height = h
	return m
}

func (m HelpModel) SetBindings(b []key.Binding) HelpModel {
	m.bindings = b
	return m
}

func (m HelpModel) Visible() bool {
	return m.visible
}

func (m HelpModel) Toggle() HelpModel {
	m.visible = !m.visible
	return m
}

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "?" || keyMsg.String() == "esc" || keyMsg.String() == "q" {
			m.visible = false
		}
	}
	return m, nil
}

func (m HelpModel) View() string {
	if !m.visible {
		return ""
	}

	title := TitleStyle.Render("Keybindings")

	var lines []string
	for _, b := range m.bindings {
		help := b.Help()
		line := lipgloss.NewStyle().
			Width(12).
			Foreground(ColorSecondary).
			Render(help.Key) +
			"  " + help.Desc
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	box := BorderFocused.
		Width(40).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", content))

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}
