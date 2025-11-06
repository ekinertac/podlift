package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestSuccessMessage(t *testing.T) {
	msg := Success("Test message")
	if !strings.Contains(msg, "Test message") {
		t.Errorf("Success() should contain message, got: %v", msg)
	}
	if !strings.Contains(msg, SymbolCheck) {
		t.Errorf("Success() should contain check symbol, got: %v", msg)
	}
}

func TestErrorMessage(t *testing.T) {
	msg := Error("Error message")
	if !strings.Contains(msg, "Error message") {
		t.Errorf("Error() should contain message, got: %v", msg)
	}
	if !strings.Contains(msg, SymbolCross) {
		t.Errorf("Error() should contain cross symbol, got: %v", msg)
	}
}

func TestStepList(t *testing.T) {
	steps := NewStepList([]string{"Step 1", "Step 2", "Step 3"})

	if len(steps.Steps) != 3 {
		t.Errorf("NewStepList() created %d steps, want 3", len(steps.Steps))
	}

	// All should be pending initially
	for i, step := range steps.Steps {
		if step.Status != "pending" {
			t.Errorf("Step %d status = %v, want pending", i, step.Status)
		}
	}

	// Start first step
	steps.Start(0, "Running step 1")
	if steps.Steps[0].Status != "running" {
		t.Errorf("After Start(0), status = %v, want running", steps.Steps[0].Status)
	}

	// Complete first step
	steps.Complete(0, "Step 1 done")
	if steps.Steps[0].Status != "done" {
		t.Errorf("After Complete(0), status = %v, want done", steps.Steps[0].Status)
	}

	// Fail second step
	steps.Fail(1, "Step 2 failed")
	if steps.Steps[1].Status != "error" {
		t.Errorf("After Fail(1), status = %v, want error", steps.Steps[1].Status)
	}

	// Render should not crash
	output := steps.Render()
	if output == "" {
		t.Error("Render() returned empty string")
	}

	// Should contain step names
	if !strings.Contains(output, "Step 1") {
		t.Error("Render() should contain 'Step 1'")
	}
}

func TestProgressBar(t *testing.T) {
	bar := NewProgressBar(40)
	if bar == nil {
		t.Fatal("NewProgressBar() returned nil")
	}

	// Render at 0%
	output := bar.Render(0.0, "Starting")
	if !strings.Contains(output, "Starting") {
		t.Errorf("Render() should contain label, got: %v", output)
	}

	// Render at 50 percent
	output = bar.Render(0.5, "Halfway")
	if !strings.Contains(output, "Halfway") {
		t.Errorf("Render() should contain label, got: %v", output)
	}
	if !strings.Contains(output, "50") {
		t.Errorf("Render() should contain percentage, got: %v", output)
	}

	// Render at 100 percent
	output = bar.Render(1.0, "Complete")
	if !strings.Contains(output, "100") {
		t.Errorf("Render() should contain 100 percent, got: %v", output)
	}
}

func TestFormatServiceTable(t *testing.T) {
	services := [][]string{
		{"web", "2/2", "healthy", "abc123", "5m"},
		{"worker", "1/1", "healthy", "abc123", "5m"},
	}

	output := FormatServiceTable(services)
	if output == "" {
		t.Error("FormatServiceTable() returned empty string")
	}

	// Should contain service names
	if !strings.Contains(output, "web") {
		t.Error("FormatServiceTable() should contain 'web'")
	}
	if !strings.Contains(output, "worker") {
		t.Error("FormatServiceTable() should contain 'worker'")
	}
}

func TestNewTable(t *testing.T) {
	columns := []table.Column{
		{Title: "Name", Width: 10},
		{Title: "Value", Width: 10},
	}
	rows := []table.Row{
		{"key1", "value1"},
		{"key2", "value2"},
	}

	tbl := NewTable(columns, rows)
	if tbl == nil {
		t.Fatal("NewTable() returned nil")
	}

	output := tbl.Render()
	if output == "" {
		t.Error("Table Render() returned empty string")
	}
}

