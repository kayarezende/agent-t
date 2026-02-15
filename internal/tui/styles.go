package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("63")).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 2).
			MarginBottom(1)

	stepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Bold(true).
			MarginBottom(1)

	selectionLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Width(10)

	selectionValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("120")).
				Bold(true)

	confirmBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 3).
			MarginTop(1).
			MarginBottom(1)

	confirmLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Width(12)

	confirmValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Underline(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

func newStyledDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("230")).
		BorderLeftForeground(lipgloss.Color("63"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("120")).
		BorderLeftForeground(lipgloss.Color("63"))
	return d
}
