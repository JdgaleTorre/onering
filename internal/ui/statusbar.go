package ui

import "github.com/charmbracelet/lipgloss"

type StatusBarModel struct {
	mode  string
	width int
}

func NewStatusBarModel() StatusBarModel {
	return StatusBarModel{mode: "NORMAL"}
}

func (m StatusBarModel) SetWidth(w int) StatusBarModel {
	m.width = w
	return m
}

func (m StatusBarModel) SetMode(mode string) StatusBarModel {
	m.mode = mode
	return m
}

func (m StatusBarModel) View() string {
	modeStyle := ModeNormalStyle
	if m.mode == "INSERT" {
		modeStyle = ModeInsertStyle
	}

	mode := modeStyle.Render(m.mode)
	hints := MutedStyle.Render(" ? help  q quit")

	left := mode + hints
	gap := m.width - lipgloss.Width(left)
	if gap < 0 {
		gap = 0
	}

	return StatusBarStyle.Width(m.width).Render(
		left + lipgloss.NewStyle().Width(gap).Render(""),
	)
}
