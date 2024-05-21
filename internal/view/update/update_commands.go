package update

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Commands struct {
	commands                     []tea.Cmd
	InputCommand                 tea.Cmd
	SpinnerCommand               tea.Cmd
	PlayersCommand               tea.Cmd
	IssueViewCommand             tea.Cmd
	IssuesListViewCommand        tea.Cmd
	GameEventHandlerCommand      tea.Cmd
	TransportEventHandlerCommand tea.Cmd
}

func NewUpdateCommands() *Commands {
	return &Commands{
		commands: make([]tea.Cmd, 0, 8),
	}
}

func (u *Commands) AppendCommand(command tea.Cmd) {
	u.commands = append(u.commands, command)
}

func (u *Commands) AppendMessage(message tea.Msg) {
	u.commands = append(u.commands, func() tea.Msg {
		return message
	})
}

func (u *Commands) Batch() tea.Cmd {
	u.commands = append(u.commands,
		u.InputCommand,
		u.SpinnerCommand,
		u.PlayersCommand,
		u.IssueViewCommand,
		u.IssuesListViewCommand,
		u.GameEventHandlerCommand,
		u.TransportEventHandlerCommand,
	)
	return tea.Batch(u.commands...)
}
