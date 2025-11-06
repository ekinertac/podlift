package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// ProgressBar represents a simple progress bar
type ProgressBar struct {
	model progress.Model
	width int
}

// NewProgressBar creates a new progress bar
func NewProgressBar(width int) *ProgressBar {
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(width),
		progress.WithoutPercentage(),
	)

	return &ProgressBar{
		model: prog,
		width: width,
	}
}

// Render renders the progress bar at given percentage (0.0 - 1.0)
func (p *ProgressBar) Render(percent float64, label string) string {
	bar := p.model.ViewAs(percent)
	
	percentStr := fmt.Sprintf("%3.0f%%", percent*100)
	return fmt.Sprintf("%s %s %s", label, bar, percentStr)
}

// Step represents a deployment step
type Step struct {
	Name     string
	Status   string // "pending", "running", "done", "error"
	Message  string
}

// StepList manages multiple steps
type StepList struct {
	Steps   []Step
	Current int
}

// NewStepList creates a new step list
func NewStepList(steps []string) *StepList {
	stepList := &StepList{
		Steps:   make([]Step, len(steps)),
		Current: 0,
	}

	for i, name := range steps {
		stepList.Steps[i] = Step{
			Name:   name,
			Status: "pending",
		}
	}

	return stepList
}

// Start marks a step as running
func (s *StepList) Start(index int, message string) {
	if index >= 0 && index < len(s.Steps) {
		s.Steps[index].Status = "running"
		s.Steps[index].Message = message
		s.Current = index
	}
}

// Complete marks a step as done
func (s *StepList) Complete(index int, message string) {
	if index >= 0 && index < len(s.Steps) {
		s.Steps[index].Status = "done"
		s.Steps[index].Message = message
	}
}

// Fail marks a step as failed
func (s *StepList) Fail(index int, message string) {
	if index >= 0 && index < len(s.Steps) {
		s.Steps[index].Status = "error"
		s.Steps[index].Message = message
	}
}

// Render renders the step list
func (s *StepList) Render() string {
	var lines []string

	for i, step := range s.Steps {
		var symbol string
		var style lipgloss.Style

		switch step.Status {
		case "done":
			symbol = SymbolCheck
			style = StyleSuccess
		case "error":
			symbol = SymbolCross
			style = StyleError
		case "running":
			symbol = "→"
			style = StyleProgress
		default: // pending
			symbol = " "
			style = StyleInfo
		}

		prefix := fmt.Sprintf("[%d/%d]", i+1, len(s.Steps))
		line := fmt.Sprintf("%s %s %s", prefix, style.Render(symbol), step.Name)
		
		if step.Message != "" {
			line += "\n    " + StyleInfo.Render(step.Message)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// RenderCurrent renders only the current step with context
func (s *StepList) RenderCurrent() string {
	if s.Current >= len(s.Steps) {
		return ""
	}

	step := s.Steps[s.Current]
	prefix := fmt.Sprintf("[%d/%d]", s.Current+1, len(s.Steps))
	
	var symbol string
	switch step.Status {
	case "running":
		symbol = "→"
	default:
		symbol = " "
	}

	line := fmt.Sprintf("%s %s %s", prefix, StyleProgress.Render(symbol), step.Name)
	if step.Message != "" {
		line += "\n  " + StyleInfo.Render(step.Message)
	}

	return line
}

