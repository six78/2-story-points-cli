package commands

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	// Common
	ToggleView  key.Binding
	ToggleInput key.Binding
	// Issues list
	NextIssue     key.Binding
	PreviousIssue key.Binding
	SelectIssue   key.Binding
	// Deck view
	NextCard     key.Binding
	PreviousCard key.Binding
	SelectCard   key.Binding
	// Dealer controls
	RevealVotes key.Binding
	FinishVote  key.Binding
	AddIssue    key.Binding
	// Room controls
	NewRoom  key.Binding
	JoinRoom key.Binding
	ExitRoom key.Binding
}

var DefaultKeyMap = KeyMap{
	// Common
	ToggleView: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("Tab", "Toggle room view"),
	),
	ToggleInput: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("Shift+Tab", "Toggle input mode"),
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
	// Dealer controls
	RevealVotes: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("R", "Reveal votes")),
	FinishVote: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("F", "Finish vote and deal next issue")),
	AddIssue: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("A", "Add issue")),
	// Room join controls
	NewRoom: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("N", "Create a new room")),
	JoinRoom: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("J", "Join room")),
	ExitRoom: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("E", "Exit room"),
	),
}
