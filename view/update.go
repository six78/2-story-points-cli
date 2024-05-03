package view

import (
	tea "github.com/charmbracelet/bubbletea"
)

type UpdateCommands struct {
	commands              []tea.Cmd
	InputCommand          tea.Cmd
	SpinnerCommand        tea.Cmd
	PlayersCommand        tea.Cmd
	IssueViewCommand      tea.Cmd
	IssuesListViewCommand tea.Cmd
}

func NewUpdateCommands() *UpdateCommands {
	return &UpdateCommands{
		InputCommand: nil,
		commands:     make([]tea.Cmd, 0, 8),
	}
}

func (u *UpdateCommands) AppendCommand(command tea.Cmd) {
	u.commands = append(u.commands, command)
}

func (u *UpdateCommands) AppendMessage(message tea.Msg) {
	u.commands = append(u.commands, func() tea.Msg {
		return message
	})
}

func (u *UpdateCommands) Batch() tea.Cmd {
	u.commands = append(u.commands,
		u.InputCommand,
		u.SpinnerCommand,
		u.PlayersCommand,
		u.IssueViewCommand,
		u.IssuesListViewCommand,
	)
	return tea.Batch(u.commands...)
}
