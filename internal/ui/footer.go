package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func padRight(s string, width int) string {
	if width <= 0 {
		return s
	}
	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

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
	var left string
	rightLine1 := FooterDimStyle.Render(data.Model)
	rightLine2 := FooterDimStyle.Render(fmt.Sprintf("%d tokens", data.TotalTokens))
	rightWidth := lipgloss.Width(rightLine1)
	if w2 := lipgloss.Width(rightLine2); w2 > rightWidth {
		rightWidth = w2
	}

	if data.Streaming {
		parts := []string{
			FooterKeyStyle.Render("ctrl+c") + FooterDimStyle.Render(" cancel"),
		}
		if data.InThinking {
			parts = append(parts, FooterKeyStyle.Render("ctrl+e")+FooterDimStyle.Render(" thinking"))
		}
		left = strings.Join(parts, "  ")

		// While streaming, keep model visible and render the live token count in white
		// right before the model label.
		rightLine1 = FooterValueStyle.Render(fmt.Sprintf("%.1f  tok/s", data.TokPerSec)) + " " + FooterDimStyle.Render(data.Model)
		rightLine2 = ""
		if data.TokPerSec > 0 {
			rightLine2 = FooterDimStyle.Render(fmt.Sprintf("%d tokens", data.TotalTokens))
		}

		rightWidth = lipgloss.Width(rightLine1)
		if w2 := lipgloss.Width(rightLine2); w2 > rightWidth {
			rightWidth = w2
		}
	} else {
		left = FooterKeyStyle.Render("/") + FooterDimStyle.Render("cmd") + "  " +
			FooterKeyStyle.Render("ctrl+h") + FooterDimStyle.Render(" help")
	}

	gap := width - lipgloss.Width(left) - rightWidth - 2
	if gap < 1 {
		gap = 1
	}

	line1 := padRight(left+strings.Repeat(" ", gap)+rightLine1, width)
	line2 := padRight(strings.Repeat(" ", lipgloss.Width(left)+gap)+rightLine2, width)
	return FooterStyle.Width(width).Render(line1 + "\n" + line2)
}
