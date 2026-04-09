package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

var mdRenderer *glamour.TermRenderer

func init() {
	var err error
	mdRenderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0), // we control width ourselves
	)
	if err != nil {
		mdRenderer = nil
	}
}

// RenderMarkdown renders markdown text to styled terminal output.
// Falls back to plain text if renderer is unavailable.
func RenderMarkdown(text string, width int) string {
	if mdRenderer == nil || text == "" {
		return text
	}

	rendered, err := mdRenderer.Render(text)
	if err != nil {
		return text
	}

	// glamour adds trailing newlines, trim them
	return strings.TrimRight(rendered, "\n")
}
