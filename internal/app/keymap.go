package app

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Send           key.Binding
	Cancel         key.Binding
	Help           key.Binding
	ExpandThinking key.Binding
	Save           key.Binding
	Load           key.Binding
	Quit           key.Binding
	Up             key.Binding
	Down           key.Binding
	Escape         key.Binding
	Tab            key.Binding
}

var keys = keyMap{
	Send: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send message"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "cancel / exit"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle help"),
	),
	ExpandThinking: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "expand/collapse thinking"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save session"),
	),
	Load: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "load session"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "previous"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "next"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "complete"),
	),
}
