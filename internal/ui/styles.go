// styles for lean

package ui

import "github.com/charmbracelet/lipgloss"

var (
	Accent  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	Warning = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Err     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	Muted   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	Bold    = lipgloss.NewStyle().Bold(true)
	Active  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	Banner  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
)

func Bolt() string {
	return Accent.Render("⚡")
}

func Ok(msg string) string {
	return Bolt() + " " + Success.Render(msg)
}

func Warn(msg string) string {
	return Bolt() + " " + Warning.Render(msg)
}

func Fail(msg string) string {
	return Bolt() + " " + Err.Render(msg)
}

func Info(msg string) string {
	return Bolt() + " " + msg
}

func Faint(msg string) string {
	return Muted.Render(msg)
}