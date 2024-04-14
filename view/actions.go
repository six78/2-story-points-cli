package view

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/exp/slices"
	"strconv"
	"strings"
	"waku-poker-planning/game"
	"waku-poker-planning/protocol"
)

type Action string

const (
	Rename Action = "rename"
	New    Action = "new"
	Join   Action = "join"
	Vote   Action = "vote"
	Deal   Action = "deal"
	Add    Action = "add"
	Reveal Action = "reveal"
	Finish Action = "finish"
	Deck   Action = "deck"
	Select Action = "select"
)

var actions = map[Action]func(m *model, args []string) (tea.Cmd, error){
	Rename: runRenameAction,
	Vote:   runVoteAction,
	Deal:   runDealAction,
	Add:    runAddAction,
	New:    runNewAction,
	Join:   runJoinAction,
	Reveal: runRevealAction,
	Finish: runFinishAction,
	Deck:   runDeckAction,
	Select: runSelectAction,
}

// FIXME: actions should be fully in tea.Cmd.
// 		  No error can be returned from here.

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

func parseVote(input string) (protocol.VoteValue, error) {
	return protocol.VoteValue(input), nil
}

func runVoteAction(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty vote")
	}
	vote, err := parseVote(args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse vote: %w", err)
	}
	cmd := func() tea.Msg {
		err := m.app.Game.PublishVote(vote)
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
		_, err := m.app.Game.Deal(args[0])
		if err != nil {
			return ActionErrorMessage{err}
		}
		return nil
	}
	return cmd, nil
}

func runAddAction(m *model, args []string) (tea.Cmd, error) {
	cmd := func() tea.Msg {
		if len(args) == 0 {
			return ActionErrorMessage{err: errors.New("empty issue")}
		}
		_, err := m.app.Game.AddIssue(args[0])
		if err != nil {
			return ActionErrorMessage{err}
		}
		return nil
	}
	return cmd, nil
}

func runNewAction(m *model, args []string) (tea.Cmd, error) {
	m.state = CreatingRoom
	return createNewRoom(m.app), nil
}

func runJoinAction(m *model, args []string) (tea.Cmd, error) {
	if len(args) == 0 {
		return nil, errors.New("empty room ID")
	}
	m.state = JoiningRoom
	return joinRoom(args[0], m.app), nil
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
	cmd := func() tea.Msg {
		if len(args) == 0 {
			return ActionErrorMessage{err: errors.New("empty result")}
		}
		result, err := parseVote(args[0])
		if err != nil {
			return ActionErrorMessage{err}
		}
		if !slices.Contains(m.gameState.Deck, result) {
			return ActionErrorMessage{err: errors.New("result not in deck")}
		}
		err = m.app.Game.Finish(result)
		if err != nil {
			return ActionErrorMessage{err}
		}
		return nil
	}
	return cmd, nil
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
		if _, ok := cards[card]; ok {
			return nil, fmt.Errorf("duplicate card: '%s'", card)
		}
		cards[card] = struct{}{}
		deck = append(deck, protocol.VoteValue(card))
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

func runSelectAction(m *model, args []string) (tea.Cmd, error) {
	cmd := func() tea.Msg {
		if len(args) == 0 {
			return ActionErrorMessage{err: errors.New("no issue index given")}
		}

		index, err := strconv.Atoi(args[0])
		if err != nil {
			return ActionErrorMessage{err: fmt.Errorf("invalid issue index: %s", args[0])}
		}

		err = m.app.Game.SelectIssue(index)
		if err != nil {
			return ActionErrorMessage{err}
		}

		return nil
	}
	return cmd, nil
}
