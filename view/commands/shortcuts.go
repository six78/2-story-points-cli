package commands

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	// Common
	ToggleView  key.Binding
	CommandMode key.Binding
	LeaveRoom   key.Binding
	// Issues list
	NextIssue     key.Binding
	PreviousIssue key.Binding
	SelectIssue   key.Binding
	// Deck view
	NextCard     key.Binding
	PreviousCard key.Binding
	SelectCard   key.Binding
}

var DefaultKeyMap = KeyMap{
	// Common
	ToggleView: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("Tab", "Toggle room view"),
	),
	CommandMode: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("C", "Switch to command mode"),
	),
	LeaveRoom: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("Q", "Leave room"),
	),
	// Issues list
	NextIssue: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "Next issue"),
	),
	PreviousIssue: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "Previous issue"),
	),
	SelectIssue: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "Select issue"),
	),
	// Deck view
	NextCard: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "Next issue"),
	),
	PreviousCard: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "Previous issue"),
	),
	SelectCard: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "Select card"),
	),
}
