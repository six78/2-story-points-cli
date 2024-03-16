package view

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"strings"
	"time"
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
	Sleep         = "sleep"
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
	Sleep:  processSleep,
}

// FIXME: actions should be fully in tea.Cmd.
// 		  No error can be returned from here.

func processSleep(m *model, args []string) (tea.Cmd, error) {
	cmd := func() tea.Msg {
		time.Sleep(5 * time.Second)
		return nil
	}
	return cmd, nil
}

func processPlayerNameInput(m *model, playerName string) (tea.Cmd, error) {
	cmd := func() tea.Msg {
		m.app.Game.RenamePlayer(playerName)
		return AppStateMessage{finishedState: InputPlayerName}
	}
	return cmd, nil
}

func runRenameAction(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty user")
	}
	cmd := func() tea.Msg {
		m.app.Game.RenamePlayer(args[0])
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
		err := m.app.Game.Deal(args[0])
		if err != nil {
			return ActionErrorMessage{err}
		}
		return nil
	}
	return cmd, nil
}

func runNewAction(m *model, args []string) (tea.Cmd, error) {
	m.state = CreatingSession
	return createNewSession(m.app), nil
}

func runJoinAction(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty session ID")
	}
	m.state = JoiningSession
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
