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

type Action string

const (
	Rename Action = "rename"
	Vote          = "vote"
	Deal          = "deal"
	New           = "new"
	Join          = "join"
)

var actions = map[Action]func(m *model, args []string) (tea.Cmd, error){
	Rename: runRenameAction,
	Vote:   runVoteAction,
	Deal:   runDealAction,
	New:    runNewAction,
	Join:   runJoinAction,
}

func runAction(m *model, command string) tea.Cmd {
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

	commandRoot := Action(args[0])
	commandFn, ok := actions[commandRoot]
	if !ok {
		err = fmt.Errorf("unknown command: %s", commandRoot)
		return nil
	}

	cmd, err := commandFn(m, args[1:])

	return cmd
}

func runRenameAction(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty user")
	}
	cmd := func() tea.Msg {
		m.app.RenamePlayer(args[0])
		return nil
	}
	return cmd, nil
}

func runVoteAction(m *model, args []string) (tea.Cmd, error) {
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

func runDealAction(m *model, args []string) (tea.Cmd, error) {
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

func runNewAction(m *model, args []string) (tea.Cmd, error) {
	m.appState = app.CreatingSession
	return createNewSession(m.app), nil
}

func runJoinAction(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty session ID")
	}
	m.appState = app.JoiningSession
	return joinSession(args[0], m.app), nil
}
