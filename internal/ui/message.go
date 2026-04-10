package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"rig-chat/internal/config"
)

const (
	assistantLeftInset = 2
	userLeftInset      = 6
)

// RenderMessage renders a single chat message for the viewport
func RenderMessage(msg config.DisplayMessage, width int, thinkingExpanded bool) string {
	var b strings.Builder
	b.WriteString("\n")

	inset := assistantLeftInset
	if msg.Role == "user" {
		inset = userLeftInset
	}
	prefix := strings.Repeat(" ", inset)

	bubbleWidth := width - inset
	if bubbleWidth < 20 {
		bubbleWidth = 20
	}

	// Header line: date left, metadata right
	header := renderHeader(msg, bubbleWidth)
	b.WriteString(prefix)
	b.WriteString(header)
	b.WriteString("\n")

	// Message body
	style := AssistantMsgStyle
	if msg.Role == "user" {
		style = UserMsgStyle
	}

	bodyWidth := bubbleWidth - 2 // padding
	if bodyWidth < 20 {
		bodyWidth = 20
	}
	style = style.Width(bodyWidth)

	body := msg.Text
	if body == "" && msg.Role == "assistant" {
		body = "..."
	}

	// Keep plain text rendering in chat bubbles to preserve consistent
	// full-width backgrounds. ANSI sequences from markdown rendering can
	// reset terminal background mid-line and create visual striping.

	b.WriteString(prefix)
	b.WriteString(style.Render(body))

	// Thinking block (collapsed/expanded)
	if msg.ThinkingText != "" {
		b.WriteString("\n")
		if thinkingExpanded {
			thinkStyle := ThinkingStyle.Width(bodyWidth)
			b.WriteString(prefix)
			b.WriteString(ThinkingLabelStyle.Render("  thinking"))
			b.WriteString("\n")
			b.WriteString(prefix)
			b.WriteString(thinkStyle.Render(msg.ThinkingText))
		} else {
			lines := strings.Count(msg.ThinkingText, "\n") + 1
			label := fmt.Sprintf("  thinking (%d lines, ctrl+e to expand)", lines)
			b.WriteString(prefix)
			b.WriteString(ThinkingLabelStyle.Render(label))
		}
	}

	b.WriteString("\n")
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
	if msg.ResponseTimeMs > 0 {
		right = append(right, formatDuration(msg.ResponseTimeMs))
	}
	totalTok := msg.InputTokens + msg.OutputTokens
	if totalTok > 0 {
		right = append(right, fmt.Sprintf("%d tokens", totalTok))
	}

	rightStr := strings.Join(right, "  ")
	leftStr := date

	gap := width - lipgloss.Width(leftStr) - lipgloss.Width(rightStr) - 2
	if gap < 1 {
		gap = 1
	}

	header := leftStr + strings.Repeat(" ", gap) + rightStr
	style := AssistantHeaderStyle
	if msg.Role == "user" {
		style = UserHeaderStyle
	}
	return style.Width(width).Render(header)
}

// RenderStreamingMessage renders the in-progress streaming message
func RenderStreamingMessage(text, thinkingText string, inThinking bool, width int, createdAt time.Time) string {
	var b strings.Builder
	b.WriteString("\n")

	prefix := strings.Repeat(" ", assistantLeftInset)
	bubbleWidth := width - assistantLeftInset
	if bubbleWidth < 20 {
		bubbleWidth = 20
	}

	bodyWidth := bubbleWidth - 2
	if bodyWidth < 20 {
		bodyWidth = 20
	}

	streamHeader := renderStreamingHeader(createdAt, bubbleWidth)
	b.WriteString(prefix)
	b.WriteString(streamHeader)
	b.WriteString("\n")

	if inThinking && text == "" {
		// Still in thinking phase, show spinner
		b.WriteString(prefix)
		b.WriteString(ThinkingLabelStyle.Render("  thinking..."))
		b.WriteString("\n")
	}

	if text != "" {
		style := AssistantMsgStyle.Width(bodyWidth)
		b.WriteString(prefix)
		b.WriteString(style.Render(text))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	return b.String()
}

func renderStreamingHeader(createdAt time.Time, width int) string {
	leftStr := createdAt.Format("15:04:05")
	return AssistantHeaderStyle.Width(width).Render(leftStr)
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	d := time.Duration(ms) * time.Millisecond
	if d < time.Minute {
		return fmt.Sprintf("%.1f sec", d.Seconds())
	}
	minutes := int(d / time.Minute)
	seconds := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}
