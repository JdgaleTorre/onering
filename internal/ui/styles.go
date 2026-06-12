package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSecondary = lipgloss.Color("#A78BFA")
	ColorBorder    = lipgloss.Color("#374151")
	ColorFocused   = lipgloss.Color("#7C3AED")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorText      = lipgloss.Color("#F9FAFB")
	ColorInsert    = lipgloss.Color("#10B981")
	ColorError     = lipgloss.Color("#EF4444")
	ColorRunning   = lipgloss.Color("#F59E0B")

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
			Background(lipgloss.Color("#1F2937")).
			Foreground(ColorText)

	ModeNormalStyle = lipgloss.NewStyle().
			Bold(true).
			Background(ColorPrimary).
			Foreground(ColorText).
			Padding(0, 1)

	ModeInsertStyle = lipgloss.NewStyle().
			Bold(true).
			Background(ColorInsert).
			Foreground(ColorText).
			Padding(0, 1)
)
