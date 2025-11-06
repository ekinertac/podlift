package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Table creates a formatted table
type Table struct {
	model table.Model
}

// NewTable creates a new table with given columns and rows
func NewTable(columns []table.Column, rows []table.Row) *Table {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorPrimary).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(ColorPrimary).
		Bold(false)
	t.SetStyles(s)

	return &Table{model: t}
}

// Render renders the table
func (t *Table) Render() string {
	return t.model.View()
}

// FormatServerTable creates a formatted server status table
func FormatServerTable(servers [][]string) string {
	if len(servers) == 0 {
		return StyleInfo.Render("No servers configured")
	}

	columns := []table.Column{
		{Title: "Server", Width: 20},
		{Title: "Role", Width: 10},
		{Title: "User", Width: 10},
		{Title: "Status", Width: 10},
	}

	rows := make([]table.Row, len(servers))
	for i, server := range servers {
		rows[i] = table.Row(server)
	}

	t := NewTable(columns, rows)
	return t.Render()
}

// FormatServiceTable creates a formatted service status table
func FormatServiceTable(services [][]string) string {
	if len(services) == 0 {
		return StyleInfo.Render("No services running")
	}

	columns := []table.Column{
		{Title: "Service", Width: 15},
		{Title: "Replicas", Width: 10},
		{Title: "Status", Width: 10},
		{Title: "Version", Width: 10},
		{Title: "Uptime", Width: 15},
	}

	rows := make([]table.Row, len(services))
	for i, service := range services {
		rows[i] = table.Row(service)
	}

	t := NewTable(columns, rows)
	return t.Render()
}

