package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FooterData holds dynamic footer information
type FooterData struct {
	Model       string
	TotalTokens int
	TokPerSec   float64
	Streaming   bool
	InThinking  bool
}

// RenderFooter renders the fixed footer bar
func RenderFooter(data FooterData, width int) string {
	var left, right string

	if data.Streaming {
		parts := []string{
			FooterKeyStyle.Render("ctrl+c") + FooterDimStyle.Render(" cancel"),
		}
		if data.InThinking {
			parts = append(parts, FooterKeyStyle.Render("ctrl+e")+FooterDimStyle.Render(" thinking"))
		}
		left = strings.Join(parts, "  ")

		rightParts := []string{}
		if data.TokPerSec > 0 {
			rightParts = append(rightParts, fmt.Sprintf("%.1f tok/s", data.TokPerSec))
		}
		rightParts = append(rightParts, FooterDimStyle.Render(data.Model))
		if data.TotalTokens > 0 {
			rightParts = append(rightParts, fmt.Sprintf("%dt", data.TotalTokens))
		}
		right = strings.Join(rightParts, "  ")
	} else {
		left = FooterDimStyle.Render("rig-chat") + "  " +
			FooterKeyStyle.Render("/") + FooterDimStyle.Render("cmd") + "  " +
			FooterKeyStyle.Render("ctrl+h") + FooterDimStyle.Render(" help")

		rightParts := []string{FooterDimStyle.Render(data.Model)}
		if data.TotalTokens > 0 {
			rightParts = append(rightParts, fmt.Sprintf("%dt", data.TotalTokens))
		}
		right = strings.Join(rightParts, "  ")
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	line := left + strings.Repeat(" ", gap) + right
	return FooterStyle.Width(width).Render(line)
}
