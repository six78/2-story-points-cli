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
	Deal               = "deal"
	New                = "new"
	Join               = "join"
)

var userCommands = map[UserCommand]func(m *model, args []string) (tea.Cmd, error){
	Online: processOnlineCommand,
	Rename: processRenameCommand,
	Vote:   processVoteCommand,
	Deal:   processDealCommand,
	New:    processNewCommand,
	Join:   processJoinCommand,
}

func processUserCommand(m *model) (tea.Cmd, error) {
	command := m.input.Value()
	defer m.input.Reset()

	args := strings.Fields(command)
	if len(args) == 0 {
		return nil, errors.New("empty command")
	}

	commandRoot := UserCommand(args[0])
	commandFn, ok := userCommands[commandRoot]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", commandRoot)
	}

	cmd, err := commandFn(m, args[1:])

	return cmd, err
}

func processOnlineCommand(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty user")
	}
	cmd := func() tea.Msg {
		m.app.PublishUserOnline(args[0])
		return nil
	}
	return cmd, nil
}

func processRenameCommand(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty user")
	}
	cmd := func() tea.Msg {
		config.PlayerName = args[0]
		m.app.PublishUserOnline(config.PlayerName)
		return nil
	}
	return cmd, nil
}

func processVoteCommand(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty vote")
	}
	vote, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse vote: %w", err)
	}
	cmd := func() tea.Msg {
		m.app.PublishVote(vote)
		return nil
	}
	return cmd, nil
}

func processDealCommand(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty deal")
	}
	cmd := func() tea.Msg {
		err := m.app.Deal(args[0])
		if err != nil {
			return FatalErrorMessage{err}
		}
		return nil
	}
	return cmd, nil
}

func processNewCommand(m *model, args []string) (tea.Cmd, error) {
	m.appState = app.CreatingSession
	return createNewSession(m.app), nil
}

func processJoinCommand(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty session ID")
	}
	m.appState = app.JoiningSession
	return joinSession(args[0], m.app), nil
}
