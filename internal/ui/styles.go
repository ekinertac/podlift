package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions
var (
	ColorPrimary   = lipgloss.Color("39")  // Blue
	ColorSuccess   = lipgloss.Color("42")  // Green
	ColorError     = lipgloss.Color("196") // Red
	ColorWarning   = lipgloss.Color("214") // Orange
	ColorInfo      = lipgloss.Color("246") // Gray
	ColorAccent    = lipgloss.Color("213") // Pink
)

// Styles for consistent UI
var (
	// StyleTitle for section titles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginTop(1).
			MarginBottom(1)

	// StyleSuccess for success messages
	StyleSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	// StyleError for error messages
	StyleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorError)

	// StyleWarning for warning messages
	StyleWarning = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWarning)

	// StyleInfo for informational messages
	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorInfo)

	// StyleCommand for displaying commands
	StyleCommand = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Italic(true)

	// StyleProgress for progress indicators
	StyleProgress = lipgloss.NewStyle().
			Foreground(ColorAccent)

	// StyleBox for boxed content
	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	// StyleCode for code snippets
	StyleCode = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)

// Checkmark and cross symbols
const (
	SymbolCheck  = "✓"
	SymbolCross  = "✗"
	SymbolArrow  = "→"
	SymbolDot    = "•"
)

// Success formats a success message
func Success(msg string) string {
	return StyleSuccess.Render(SymbolCheck) + " " + msg
}

// Error formats an error message
func Error(msg string) string {
	return StyleError.Render(SymbolCross) + " " + msg
}

// Warning formats a warning message
func Warning(msg string) string {
	return StyleWarning.Render(SymbolArrow) + " " + msg
}

// Info formats an info message
func Info(msg string) string {
	return StyleInfo.Render(SymbolDot) + " " + msg
}

// Title formats a section title
func Title(msg string) string {
	return StyleTitle.Render(msg)
}

// Code formats code or command text
func Code(text string) string {
	return StyleCode.Render(text)
}

// Box wraps content in a box
func Box(content string) string {
	return StyleBox.Render(content)
}

