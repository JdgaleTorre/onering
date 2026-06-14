package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
)

type BindingGroup struct {
	Name     string
	Bindings []key.Binding
}

type ProjectInfoModel struct {
	keyBindingGroups []BindingGroup
	bannerText       string
	width            int
	height           int
}

func NewProjectInfoModel() ProjectInfoModel {
	return ProjectInfoModel{
		bannerText: figure.NewFigure("LAZYCODE", "", true).String(),
	}
}

func (m ProjectInfoModel) SetSize(w, h int) ProjectInfoModel {
	m.width = w
	m.height = h
	return m
}

func (m ProjectInfoModel) SetKeyBindingGroups(groups []BindingGroup) ProjectInfoModel {
	m.keyBindingGroups = groups
	return m
}

func (m ProjectInfoModel) SetProjectName(name string) ProjectInfoModel {
	return m
}

func (m ProjectInfoModel) View() string {
	bannerStyled := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Render(m.bannerText)

	githubURL := lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Render("https://github.com/josegale/lazycode")

	kbContent := m.renderKeybindings()

	//hint := MutedStyle.Render("Press 0 to close")

	info := lipgloss.JoinVertical(
		lipgloss.Center,
		bannerStyled,
		"",
		githubURL,
		"",
		kbContent,
		"",
		//hint,
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		info)
}

func (m ProjectInfoModel) renderKeybindings() string {
	if len(m.keyBindingGroups) == 0 {
		return ""
	}

	var leftGroups, rightGroups []BindingGroup
	for i, g := range m.keyBindingGroups {
		if i%2 == 0 {
			leftGroups = append(leftGroups, g)
		} else {
			rightGroups = append(rightGroups, g)
		}
	}

	colGap := 6

	leftContent := renderGroupColumn(leftGroups)
	rightContent := renderGroupColumn(rightGroups)

	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)

	leftPadded := lipgloss.NewStyle().Width(leftWidth).Render(leftContent)
	rightBlock := lipgloss.NewStyle().Width(rightWidth).Render(rightContent)

	leftParts := strings.Split(leftPadded, "\n")
	rightParts := strings.Split(rightBlock, "\n")

	maxLines := max(len(leftParts), len(rightParts))
	var lines []string
	for i := 0; i < maxLines; i++ {
		l := ""
		if i < len(leftParts) {
			l = leftParts[i]
		} else {
			l = strings.Repeat(" ", leftWidth)
		}
		r := ""
		if i < len(rightParts) {
			r = rightParts[i]
		}
		lines = append(lines, l+strings.Repeat(" ", colGap)+r)
	}

	return strings.Join(lines, "\n")
}

func renderGroupColumn(groups []BindingGroup) string {
	var parts []string
	for _, g := range groups {
		header := TitleStyle.Render(strings.ToUpper(g.Name))
		var bindings []string
		for _, b := range g.Bindings {
			h := b.Help()
			line := lipgloss.NewStyle().Width(10).Foreground(ColorSecondary).Render(h.Key) +
				" " + MutedStyle.Render(h.Desc)
			bindings = append(bindings, line)
		}
		section := lipgloss.JoinVertical(lipgloss.Left, header, "")
		if len(bindings) > 0 {
			section = lipgloss.JoinVertical(lipgloss.Left, header, "", strings.Join(bindings, "\n"))
		}
		parts = append(parts, section)
	}
	return strings.Join(parts, "\n\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
