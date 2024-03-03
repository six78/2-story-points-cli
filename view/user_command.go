package view

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"strings"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
)

type UserCommand string

const (
	Online UserCommand = "online"
	Rename             = "rename"
	Vote               = "vote"
)

var userCommands = map[UserCommand]func(app *app.App, args []string) (tea.Cmd, error){
	Online: processOnlineCommand,
	Rename: processRenameCommand,
	Vote:   processVoteCommand,
}

func processUserCommand(app *app.App, command string) (tea.Cmd, error) {
	args := strings.Fields(command)
	if len(args) == 0 {
		return nil, errors.New("empty command")
	}
	commandRoot := UserCommand(args[0])
	commandFn, ok := userCommands[commandRoot]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", commandRoot)
	}
	return commandFn(app, args[1:])
}

func processOnlineCommand(app *app.App, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty user")
	}
	cmd := func() tea.Msg {
		app.PublishUserOnline(args[0])
		return nil
	}
	return cmd, nil
}

func processRenameCommand(app *app.App, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty user")
	}
	cmd := func() tea.Msg {
		config.PlayerName = args[0]
		app.PublishUserOnline(config.PlayerName)
		return nil
	}
	return cmd, nil
}

func processVoteCommand(app *app.App, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty vote")
	}
	vote, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse vote: %w", err)
	}
	cmd := func() tea.Msg {
		app.PublishVote(vote)
		return nil
	}
	return cmd, nil
}
