package ui

import "github.com/charmbracelet/lipgloss"

type StatusBarModel struct {
	mode          string
	hints         string
	taskInfo      string
	version       string
	width         int
	sidebarHidden bool
}

func NewStatusBarModel() StatusBarModel {
	return StatusBarModel{
		mode:  "NORMAL",
		hints: " enter: projects  n: new session  ?: help  q: quit",
	}
}

func (m StatusBarModel) SetVersion(v string) StatusBarModel {
	m.version = v
	return m
}

func (m StatusBarModel) SetWidth(w int) StatusBarModel {
	m.width = w
	return m
}

func (m StatusBarModel) SetSidebarHidden(hidden bool) StatusBarModel {
	m.sidebarHidden = hidden
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

func (m StatusBarModel) SetTaskInfo(info string) StatusBarModel {
	m.taskInfo = info
	return m
}

func (m StatusBarModel) ClearTaskInfo() StatusBarModel {
	m.taskInfo = ""
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

	modeLabel := m.mode
	if m.sidebarHidden {
		modeLabel += " [S]"
	}
	mode := modeStyle.Render(modeLabel)

	hintsStr := m.hints
	modeWidth := lipgloss.Width(mode)
	if modeWidth+lipgloss.Width(MutedStyle.Render(hintsStr)) > m.width {
		helpStr := " ?: help"
		if modeWidth+lipgloss.Width(MutedStyle.Render(helpStr)) <= m.width {
			hintsStr = helpStr
		} else {
			hintsStr = ""
		}
	}
	hints := MutedStyle.Render(hintsStr)

	left := mode + hints

	rightText := m.taskInfo
	if rightText == "" && m.version != "" {
		rightText = "v" + m.version
	}

	if rightText != "" {
		sep := MutedStyle.Render(" │ ")
		info := MutedStyle.Render(rightText)
		gap := m.width - lipgloss.Width(left) - lipgloss.Width(sep) - lipgloss.Width(info)
		if gap < 0 {
			gap = 0
		}
		return StatusBarStyle.Width(m.width).Render(
			left + lipgloss.NewStyle().Width(gap).Render("") + sep + info,
		)
	}

	gap := m.width - lipgloss.Width(left)
	if gap < 0 {
		gap = 0
	}

	return StatusBarStyle.Width(m.width).Render(
		left + lipgloss.NewStyle().Width(gap).Render(""),
	)
}
