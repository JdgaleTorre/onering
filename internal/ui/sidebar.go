package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SidebarModel struct {
	focused bool
	width   int
	height  int
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

func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	return m, nil
}

func (m SidebarModel) View() string {
	style := BorderNormal
	if m.focused {
		style = BorderFocused
	}

	title := TitleStyle.Render("Sessions")
	content := MutedStyle.Render("No sessions yet.\nPress 'n' to create one.")

	inner := lipgloss.JoinVertical(lipgloss.Left, title, "", content)

	return style.
		Width(m.width - 2). // account for border
		Height(m.height - 2).
		Render(inner)
}
