package view

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
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

func processUserCommand(m *model, command string) tea.Cmd {
	var err error

	defer func() {
		config.Logger.Debug("user command processed",
			zap.Error(err),
			zap.Any("appState", m.appState),
		)

		if err != nil {
			m.lastCommandError = err.Error()
		} else {
			m.lastCommandError = ""
		}
	}()

	args := strings.Fields(command)
	if len(args) == 0 {
		err = errors.New("empty command")
		return nil
	}

	commandRoot := UserCommand(args[0])
	commandFn, ok := userCommands[commandRoot]
	if !ok {
		err = fmt.Errorf("unknown command: %s", commandRoot)
		return nil
	}

	cmd, err := commandFn(m, args[1:])

	return cmd
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
