package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/josegale/onering/internal/ui"
)

type shutdownItem struct {
	Name   string
	Closed bool
}

type shutdownItemClosedMsg struct {
	Index int
}

type shutdownTickMsg struct{}

func (m AppModel) startShutdown() (AppModel, tea.Cmd) {
	m.quitting = true
	m.shutdownFrame = 0

	var items []shutdownItem

	for _, s := range m.sessions {
		label := s.Label()
		if label == "" {
			label = s.AgentName()
		}
		items = append(items, shutdownItem{Name: label})
	}

	for _, a := range m.apps {
		if a.Sess != nil {
			items = append(items, shutdownItem{Name: a.Name})
		}
	}

	for i := range m.tasks {
		if m.tasks[i].Sess != nil {
			items = append(items, shutdownItem{Name: m.tasks[i].Task.Name})
		}
	}

	m.shutdownItems = items

	if len(items) == 0 {
		return m, tea.Quit
	}

	var cmds []tea.Cmd

	idx := 0
	for _, s := range m.sessions {
		sess := s
		i := idx
		cmds = append(cmds, func() tea.Msg {
			sess.Close()
			return shutdownItemClosedMsg{Index: i}
		})
		idx++
	}

	for _, a := range m.apps {
		if a.Sess != nil {
			sess := a.Sess
			i := idx
			cmds = append(cmds, func() tea.Msg {
				sess.Close()
				return shutdownItemClosedMsg{Index: i}
			})
			idx++
		}
	}

	for j := range m.tasks {
		if m.tasks[j].Sess != nil {
			sess := m.tasks[j].Sess
			proc := m.tasks[j].cmd
			i := idx
			cmds = append(cmds, func() tea.Msg {
				sess.Close()
				if proc != nil && proc.Process != nil {
					proc.Process.Kill()
				}
				return shutdownItemClosedMsg{Index: i}
			})
			idx++
		}
	}

	cmds = append(cmds, tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
		return shutdownTickMsg{}
	}))

	return m, tea.Batch(cmds...)
}

func (m AppModel) updateShutdown(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case shutdownItemClosedMsg:
		if msg.Index < len(m.shutdownItems) {
			m.shutdownItems[msg.Index].Closed = true
		}
		allDone := true
		for _, item := range m.shutdownItems {
			if !item.Closed {
				allDone = false
				break
			}
		}
		if allDone {
			return m, tea.Quit
		}
		return m, nil

	case shutdownTickMsg:
		m.shutdownFrame++
		allDone := true
		for _, item := range m.shutdownItems {
			if !item.Closed {
				allDone = false
				break
			}
		}
		if allDone {
			return m, tea.Quit
		}
		return m, tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
			return shutdownTickMsg{}
		})
	}
	return m, nil
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (m AppModel) shutdownView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.ColorPrimary).
		Render("Shutting down...")

	var lines []string
	for _, item := range m.shutdownItems {
		var icon string
		if item.Closed {
			icon = lipgloss.NewStyle().Foreground(ui.ColorInsert).Render("✓")
		} else {
			frame := spinnerFrames[m.shutdownFrame%len(spinnerFrames)]
			icon = lipgloss.NewStyle().Foreground(ui.ColorRunning).Render(frame)
		}
		line := fmt.Sprintf("  %s %s", icon, item.Name)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	box := ui.BorderFocused.
		Width(40).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", content))

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}
