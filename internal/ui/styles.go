package ui

import "github.com/charmbracelet/lipgloss"

const ModalWidth = 44

var (
	ColorPrimary   lipgloss.TerminalColor = lipgloss.ANSIColor(5)  // Magenta
	ColorSecondary lipgloss.TerminalColor = lipgloss.ANSIColor(13) // Bright Magenta
	ColorBorder    lipgloss.TerminalColor = lipgloss.ANSIColor(8)  // Bright Black
	ColorFocused   lipgloss.TerminalColor = lipgloss.ANSIColor(5)  // Magenta
	ColorMuted     lipgloss.TerminalColor = lipgloss.ANSIColor(8)  // Bright Black
	ColorText      lipgloss.TerminalColor = lipgloss.ANSIColor(7)  // White
	ColorInsert    lipgloss.TerminalColor = lipgloss.ANSIColor(2)  // Green
	ColorError     lipgloss.TerminalColor = lipgloss.ANSIColor(1)  // Red
	ColorRunning   lipgloss.TerminalColor = lipgloss.ANSIColor(3)  // Yellow

	BorderNormal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	BorderFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorFocused)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.ANSIColor(0)).
			Foreground(ColorText)

	ModeNormalStyle = lipgloss.NewStyle().
			Bold(true).
			Background(ColorPrimary).
			Foreground(lipgloss.ANSIColor(15)).
			Padding(0, 1)

	ModeInsertStyle = lipgloss.NewStyle().
			Bold(true).
			Background(ColorInsert).
			Foreground(lipgloss.ANSIColor(15)).
			Padding(0, 1)

	ModePassthroughStyle = lipgloss.NewStyle().
				Bold(true).
				Background(ColorRunning).
				Foreground(lipgloss.ANSIColor(0)).
				Padding(0, 1)

	RunningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRunning)

	MouseAppStyle = lipgloss.NewStyle().
			Bold(true).
			Background(ColorInsert).
			Foreground(lipgloss.ANSIColor(15)).
			Padding(0, 1)

	MousePTYStyle = lipgloss.NewStyle().
			Bold(true).
			Background(ColorSecondary).
			Foreground(lipgloss.ANSIColor(15)).
			Padding(0, 1)
)
