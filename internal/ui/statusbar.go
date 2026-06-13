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
	switch m.mode {
	case "INSERT":
		modeStyle = ModeInsertStyle
	case "PASSTHROUGH":
		modeStyle = ModePassthroughStyle
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
