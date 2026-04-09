package chat

import "strings"

// ThinkParser is a stateful parser that separates <think>...</think> blocks
// from streaming text. Ported from test/chat.sh process_chunk logic.
type ThinkParser struct {
	InThink bool
	carry   string
}

// ProcessResult holds the parsed output of a chunk
type ProcessResult struct {
	Text     string // visible text to display
	Thinking string // thinking text (hidden or shown depending on mode)
}

// Process takes a streaming chunk and returns separated text/thinking content.
// It buffers partial tag boundaries across calls.
func (p *ThinkParser) Process(chunk string) ProcessResult {
	var textBuf, thinkBuf strings.Builder

	// Normalize malformed close tag
	chunk = strings.ReplaceAll(chunk, "</ think>", "</think>")

	chunk = p.carry + chunk
	p.carry = ""

	for {
		if !p.InThink {
			// Strip stray closing tags
			chunk = strings.ReplaceAll(chunk, "</think>", "")

			if idx := strings.Index(chunk, "<think>"); idx >= 0 {
				before := chunk[:idx]
				after := chunk[idx+7:] // len("<think>") == 7
				textBuf.WriteString(before)
				p.InThink = true
				chunk = after
			} else {
				// Buffer last 6 chars to catch partial "<think>" across chunks
				keep := 6
				if len(chunk) > keep {
					textBuf.WriteString(chunk[:len(chunk)-keep])
					p.carry = chunk[len(chunk)-keep:]
				} else {
					p.carry = chunk
				}
				break
			}
		} else {
			if idx := strings.Index(chunk, "</think>"); idx >= 0 {
				inside := chunk[:idx]
				after := chunk[idx+8:] // len("</think>") == 8
				thinkBuf.WriteString(inside)
				p.InThink = false
				chunk = after
			} else {
				// Buffer last 7 chars to catch partial "</think>" across chunks
				keep := 7
				if len(chunk) > keep {
					thinkBuf.WriteString(chunk[:len(chunk)-keep])
					p.carry = chunk[len(chunk)-keep:]
				} else {
					p.carry = chunk
				}
				break
			}
		}
	}

	return ProcessResult{
		Text:     textBuf.String(),
		Thinking: thinkBuf.String(),
	}
}

// Flush returns any remaining buffered content
func (p *ThinkParser) Flush() ProcessResult {
	if p.carry == "" {
		return ProcessResult{}
	}

	carry := p.carry
	p.carry = ""

	// Don't emit tag fragments
	if carry == "<think>" || carry == "</think>" {
		return ProcessResult{}
	}

	if p.InThink {
		return ProcessResult{Thinking: carry}
	}
	return ProcessResult{Text: carry}
}
