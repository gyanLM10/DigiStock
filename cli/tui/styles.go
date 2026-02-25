package tui

import "github.com/charmbracelet/lipgloss"

var (
	cyan    = lipgloss.Color("#00D7FF")
	green   = lipgloss.Color("#00FF87")
	yellow  = lipgloss.Color("#FFD700")
	red     = lipgloss.Color("#FF5F5F")
	dim     = lipgloss.Color("#555555")
	white   = lipgloss.Color("#DDDDDD")
	magenta = lipgloss.Color("#FF79C6")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyan).
			Padding(0, 2)

	cursorStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(white)

	descStyle = lipgloss.NewStyle().
			Foreground(dim)

	hintStyle = lipgloss.NewStyle().
			Foreground(dim)

	labelStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(dim)
)
