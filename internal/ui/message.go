package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"rig-chat/internal/config"
)

// RenderMessage renders a single chat message for the viewport
func RenderMessage(msg config.DisplayMessage, width int, thinkingExpanded bool) string {
	var b strings.Builder

	// Header line (dim): date left, metadata right
	header := renderHeader(msg, width)
	b.WriteString(header)
	b.WriteString("\n")

	// Message body
	style := AssistantMsgStyle
	if msg.Role == "user" {
		style = UserMsgStyle
	}

	bodyWidth := width - 2 // padding
	if bodyWidth < 20 {
		bodyWidth = 20
	}
	style = style.Width(bodyWidth)

	body := msg.Text
	if body == "" && msg.Role == "assistant" {
		body = "..."
	}

	// Render markdown for assistant messages
	if msg.Role == "assistant" && body != "..." {
		rendered := RenderMarkdown(body, bodyWidth)
		if rendered != "" {
			body = rendered
		}
	}

	b.WriteString(style.Render(body))

	// Thinking block (collapsed/expanded)
	if msg.ThinkingText != "" {
		b.WriteString("\n")
		if thinkingExpanded {
			thinkStyle := ThinkingStyle.Width(bodyWidth)
			b.WriteString(ThinkingLabelStyle.Render("  thinking"))
			b.WriteString("\n")
			b.WriteString(thinkStyle.Render(msg.ThinkingText))
		} else {
			lines := strings.Count(msg.ThinkingText, "\n") + 1
			label := fmt.Sprintf("  thinking (%d lines, ctrl+e to expand)", lines)
			b.WriteString(ThinkingLabelStyle.Render(label))
		}
	}

	b.WriteString("\n")
	return b.String()
}

func renderHeader(msg config.DisplayMessage, width int) string {
	date := msg.CreatedAt.Format("15:04:05")

	var right []string
	if msg.ImagePath != "" {
		right = append(right, msg.ImagePath)
	}
	if msg.TokensPerSecond > 0 {
		right = append(right, fmt.Sprintf("%.1f tok/s", msg.TokensPerSecond))
	}
	totalTok := msg.InputTokens + msg.OutputTokens
	if totalTok > 0 {
		right = append(right, fmt.Sprintf("%d tok", totalTok))
	}

	rightStr := strings.Join(right, "  ")
	leftStr := date

	gap := width - lipgloss.Width(leftStr) - lipgloss.Width(rightStr) - 2
	if gap < 1 {
		gap = 1
	}

	header := leftStr + strings.Repeat(" ", gap) + rightStr
	return MsgHeaderStyle.Render(header)
}

// RenderStreamingMessage renders the in-progress streaming message
func RenderStreamingMessage(text, thinkingText string, inThinking bool, width int) string {
	var b strings.Builder

	bodyWidth := width - 2
	if bodyWidth < 20 {
		bodyWidth = 20
	}

	if inThinking && text == "" {
		// Still in thinking phase, show spinner
		b.WriteString(ThinkingLabelStyle.Render("  thinking..."))
		b.WriteString("\n")
	}

	if text != "" {
		style := AssistantMsgStyle.Width(bodyWidth)
		b.WriteString(style.Render(text))
		b.WriteString("\n")
	}

	return b.String()
}
