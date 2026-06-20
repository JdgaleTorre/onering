package ui

import "github.com/charmbracelet/lipgloss"

type StatusBarModel struct {
	mode          string
	hints         string
	taskInfo      string
	version       string
	updateVersion string
	width         int
	sidebarHidden bool
	mouseMode     string
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

func (m StatusBarModel) SetMouseMode(mode string) StatusBarModel {
	m.mouseMode = mode
	return m
}

func (m StatusBarModel) SetUpdateAvailable(v string) StatusBarModel {
	m.updateVersion = v
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

	if m.mouseMode != "" {
		mouseStyle := MouseAppStyle
		if m.mouseMode == "pty" {
			mouseStyle = MousePTYStyle
		}
		mode += " " + mouseStyle.Render("mouse:"+m.mouseMode)
	}

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
	updateHint := false
	if rightText == "" && m.updateVersion != "" {
		rightText = "v" + m.updateVersion + " available · U: copy update cmd"
		updateHint = true
	} else if rightText == "" && m.version != "" {
		rightText = "v" + m.version
	}

	if rightText != "" {
		sep := MutedStyle.Render(" │ ")
		var info string
		if updateHint {
			info = lipgloss.NewStyle().Foreground(ColorRunning).Render(rightText)
		} else {
			info = MutedStyle.Render(rightText)
		}
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
