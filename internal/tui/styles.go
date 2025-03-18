package tui

import "github.com/charmbracelet/lipgloss"

var (
	panelStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FAFAFA"))
	selectedPanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7D56F4"))
	listStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	selectedListStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
)
