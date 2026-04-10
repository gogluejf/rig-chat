package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"rig-chat/internal/config"
)

// RenderMessage renders a single chat message for the viewport
func RenderMessage(msg config.DisplayMessage, width int, thinkingExpanded bool) string {
	var b strings.Builder

	// Left-margin concept removed (requested): use full available width.
	bubbleWidth := width
	if bubbleWidth < 20 {
		bubbleWidth = 20
	}

	// Header line: date left, metadata right
	header := renderHeader(msg, bubbleWidth)
	b.WriteString(header)

	// Message body
	style := AssistantMsgStyle
	if msg.Role == "user" {
		style = UserMsgStyle
	}

	bodyWidth := bubbleWidth
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
	b.WriteString(style.Render("\n" + body + "\n"))

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

	// One trailing spacer line after each message block.
	b.WriteString("\n")
	return b.String()
}

func renderHeader(msg config.DisplayMessage, width int) string {
	date := msg.CreatedAt.Format("15:04:05")

	var right []string
	if msg.ImagePath != "" {
		right = append(right, filepath.Base(msg.ImagePath))
	}
	if msg.Role == "user" {
		// User messages: show input token estimate only.
		if msg.InputTokens > 0 {
			right = append(right, fmt.Sprintf("%d tokens", msg.InputTokens))
		}
	} else {
		// Assistant messages: tok/s → response time → output tokens.
		if msg.TokensPerSecond > 0 {
			right = append(right, fmt.Sprintf("%.1f tok/s", msg.TokensPerSecond))
		}
		if msg.ResponseTimeMs > 0 {
			right = append(right, formatDuration(msg.ResponseTimeMs))
		}
		if msg.OutputTokens > 0 {
			right = append(right, fmt.Sprintf("%d tokens", msg.OutputTokens))
		}
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

	return style.Width(width).Render("\n" + header + "\n")
}

// RenderStreamingMessage renders the in-progress streaming message.
// tokenCount and tokPerSec are live values; tokPerSec should be 0 until the
// first token arrives so the latency (pre-token wait) is excluded from the
// speed calculation.
func RenderStreamingMessage(text, thinkingText string, inThinking bool, width int, createdAt time.Time, tokenCount int, tokPerSec float64) string {
	var b strings.Builder

	bubbleWidth := width
	if bubbleWidth < 20 {
		bubbleWidth = 20
	}
	bodyWidth := bubbleWidth

	streamHeader := renderStreamingHeader(createdAt, tokenCount, tokPerSec, bubbleWidth)
	b.WriteString(streamHeader)
	b.WriteString("\n")

	if inThinking && text == "" {
		b.WriteString(ThinkingLabelStyle.Render("  thinking..."))
		b.WriteString("\n")
	}

	if text != "" {
		style := AssistantMsgStyle.Width(bodyWidth)
		b.WriteString(style.Render(text + "\n"))
		b.WriteString("\n")
	}
	return b.String()
}

// renderStreamingHeader mirrors renderHeader visually:
// timestamp on the left, [elapsed  tok/s  N tokens] on the right.
// The elapsed timer starts from createdAt (before the first token) so the
// pre-token latency is visible. tok/s is only shown once > 0 (caller ensures
// it stays 0 until the first token arrives).
func renderStreamingHeader(createdAt time.Time, tokenCount int, tokPerSec float64, width int) string {
	leftStr := createdAt.Format("15:04:05")

	// Order matches the completed-message header: tok/s → elapsed → tokens.
	// tok/s is suppressed until > 0 so the latency phase shows only the timer.
	var right []string
	if tokPerSec > 0 {
		right = append(right, fmt.Sprintf("%.1f tok/s", tokPerSec))
	}
	elapsed := time.Since(createdAt)
	right = append(right, formatDuration(elapsed.Milliseconds()))
	if tokenCount > 0 {
		right = append(right, fmt.Sprintf("%d tokens", tokenCount))
	}

	rightStr := strings.Join(right, "  ")
	gap := width - lipgloss.Width(leftStr) - lipgloss.Width(rightStr) - 2
	if gap < 1 {
		gap = 1
	}

	header := leftStr + strings.Repeat(" ", gap) + rightStr
	return AssistantHeaderStyle.Width(width).Render("\n" + header + "\n")
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
