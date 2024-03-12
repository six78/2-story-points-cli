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
	"waku-poker-planning/game"
	"waku-poker-planning/protocol"
)

type Action string

const (
	Rename Action = "rename"
	New           = "new"
	Join          = "join"
	Vote          = "vote"
	Deal          = "deal"
	Reveal        = "reveal"
	Finish        = "finish"
	Deck          = "deck"
)

var actions = map[Action]func(m *model, args []string) (tea.Cmd, error){
	Rename: runRenameAction,
	Vote:   runVoteAction,
	Deal:   runDealAction,
	New:    runNewAction,
	Join:   runJoinAction,
	Reveal: runRevealAction,
	Finish: runFinishAction,
	Deck:   runDeckAction,
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
		err := m.app.Game.PublishVote(protocol.VoteResult(vote))
		if err != nil {
			return ActionErrorMessage{err}
		}
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
			return ActionErrorMessage{err}
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

func runRevealAction(m *model, args []string) (tea.Cmd, error) {
	cmd := func() tea.Msg {
		err := m.app.Game.Reveal()
		if err != nil {
			return ActionErrorMessage{err}
		}
		return nil
	}
	return cmd, nil
}

func runFinishAction(m *model, args []string) (tea.Cmd, error) {
	return nil, errors.New("action not implemented")
}

func parseDeck(args []string) (protocol.Deck, error) {

	if len(args) == 0 {
		return nil, errors.New("deck can't be empty")
	}

	if len(args) == 1 {
		// attempt to parse deck by name
		deckName := strings.ToLower(args[0])
		deck, ok := game.GetDeck(deckName)
		if !ok {
			return nil, fmt.Errorf("unknown deck: '%s', available decks: %s",
				args[0], strings.Join(game.AvailableDecks(), ", "))
		}
		return deck, nil
	}

	deck := protocol.Deck{}
	cards := map[string]struct{}{}

	for _, card := range args {
		vote, err := strconv.Atoi(card)
		if err != nil {
			return nil, fmt.Errorf("failed to parse card: '%w'", err)
		}
		if _, ok := cards[card]; ok {
			return nil, fmt.Errorf("duplicate card: '%s'", card)
		}
		cards[card] = struct{}{}
		deck = append(deck, protocol.VoteResult(vote))
	}

	return deck, nil
}

func runDeckAction(m *model, args []string) (tea.Cmd, error) {
	deck, err := parseDeck(args)
	if err != nil {
		return nil, err
	}

	cmd := func() tea.Msg {
		err := m.app.Game.SetDeck(deck)
		if err != nil {
			return ActionErrorMessage{err}
		}
		return nil
	}
	return cmd, nil
}
