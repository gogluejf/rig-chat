package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Message backgrounds
	UserMsgStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")). // dark gray
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	AssistantMsgStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("233")). // near-black
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)

	// Thinking block
	ThinkingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")). // mid-gray dim
			Italic(true).
			Padding(0, 1)

	ThinkingLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Bold(true)

	// Message header (dim line with date, tokens, etc.)
	MsgHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // dim
			Padding(0, 1)

	// Footer bar
	FooterStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("245")).
			Padding(0, 1)

	FooterKeyStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("110")). // soft blue
			Bold(true)

	FooterDimStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("240"))

	// Code blocks
	CodeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("234")).
			Foreground(lipgloss.Color("228")). // yellow
			Padding(0, 1)

	// Markdown elements
	HeadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // bright white
			Bold(true)

	BulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110")) // cyan-blue

	// Spinner / status
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")) // pink

	// Error
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // red
			Bold(true)

	// Image attachment chip
	AttachmentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // orange
			Padding(0, 1)

	// Command palette
	CommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110")).
			Bold(true)

	CommandDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	CommandSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("237")).
				Foreground(lipgloss.Color("117")).
				Bold(true)
)
