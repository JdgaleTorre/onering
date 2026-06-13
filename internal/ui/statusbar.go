package ui

import "github.com/charmbracelet/lipgloss"

type StatusBarModel struct {
	mode  string
	hints string
	width int
}

func NewStatusBarModel() StatusBarModel {
	return StatusBarModel{
		mode:  "NORMAL",
		hints: " n: new session  e: editor  g: lazygit  q: quit",
	}
}

func (m StatusBarModel) SetWidth(w int) StatusBarModel {
	m.width = w
	return m
}

func (m StatusBarModel) SetMode(mode string) StatusBarModel {
	m.mode = mode
	return m
}

func (m StatusBarModel) SetHints(hints string) StatusBarModel {
	m.hints = hints
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
	hints := MutedStyle.Render(m.hints)

	left := mode + hints
	gap := m.width - lipgloss.Width(left)
	if gap < 0 {
		gap = 0
	}

	return StatusBarStyle.Width(m.width).Render(
		left + lipgloss.NewStyle().Width(gap).Render(""),
	)
}
