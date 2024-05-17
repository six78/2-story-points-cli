package view

import (
	"2sp/internal/view/commands"
	"2sp/internal/view/messages"
	"2sp/internal/view/states"
	"2sp/pkg/game"
	protocol2 "2sp/pkg/protocol"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"strconv"
	"strings"
)

type Action string

const (
	Rename  Action = "rename"
	New     Action = "new"
	Join    Action = "join"
	Exit    Action = "exit"
	Vote    Action = "vote"
	Unvote  Action = "unvote"
	Deal    Action = "deal"
	Add     Action = "add"
	Reveal  Action = "reveal"
	Finish  Action = "finish"
	Deck    Action = "deck"
	Select  Action = "select"
	Restore Action = "restore"
)

type actionFunc func(m *model, args []string) tea.Cmd

var actions = map[Action]actionFunc{
	Rename:  runRenameAction,
	Vote:    runVoteAction,
	Unvote:  runUnvoteAction,
	Deal:    runDealAction,
	Add:     runAddAction,
	New:     runNewAction,
	Join:    runJoinAction,
	Exit:    runExitAction,
	Reveal:  runRevealAction,
	Finish:  runFinishAction,
	Deck:    runDeckAction,
	Select:  runSelectAction,
	Restore: runRestoreAction,
}

func processPlayerNameInput(m *model, playerName string) tea.Cmd {
	return func() tea.Msg {
		err := m.app.RenamePlayer(playerName)
		if err != nil {
			return messages.NewErrorMessage(err)
		}
		return messages.AppStateFinishedMessage{
			State: states.InputPlayerName,
		}
	}
}

func runRenameAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("empty user")
			return messages.NewErrorMessage(err)
		}
		err := m.app.RenamePlayer(args[0])
		return messages.NewErrorMessage(err)
	}
}

func parseVote(input string) (protocol2.VoteValue, error) {
	return protocol2.VoteValue(input), nil
}

func runVoteAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("empty vote")
			return messages.NewErrorMessage(err)
		}

		vote, err := parseVote(args[0])
		if err != nil {
			err = errors.Wrap(err, "failed to parse vote")
			return messages.NewErrorMessage(err)
		}

		return commands.PublishVote(m.app, vote)()
	}
}

func runUnvoteAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		return commands.PublishVote(m.app, "")()
	}
}

func runDealAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("empty deal")
			return messages.NewErrorMessage(err)
		}
		// TODO: Find a better way of restoring empty spaces between args
		issue := strings.Join(args, " ")
		_, err := m.app.Game.Deal(issue)
		return messages.NewErrorMessage(err)
	}
}

func runAddAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("empty issue")
			return messages.NewErrorMessage(err)
		}
		return commands.AddIssue(m.app, args[0])()
	}
}

func runNewAction(m *model, args []string) tea.Cmd {
	return commands.CreateNewRoom(m.app)
}

func runJoinAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("no room id argument provided")
			return messages.NewErrorMessage(err)
		}
		roomID := protocol2.NewRoomID(args[0])
		return commands.JoinRoom(m.app, roomID, nil)()
	}
}

func runExitAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		m.app.Game.LeaveRoom()
		return messages.RoomJoin{
			RoomID:   m.app.Game.RoomID(),
			IsDealer: m.app.Game.IsDealer(),
		}
	}
}

func runRevealAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		err := m.app.Game.Reveal()
		return messages.NewErrorMessage(err)
	}
}

func runFinishAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("empty result")
			return messages.NewErrorMessage(err)
		}
		result, err := parseVote(args[0])
		if err != nil {
			return messages.NewErrorMessage(err)
		}
		if !slices.Contains(m.gameState.Deck, result) {
			err = errors.New("result not in deck")
			return messages.NewErrorMessage(err)
		}
		err = m.app.Game.Finish(result)
		return messages.NewErrorMessage(err)
	}
}

func parseDeck(args []string) (protocol2.Deck, error) {
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

	deck := protocol2.Deck{}
	cards := map[string]struct{}{}

	for _, card := range args {
		if _, ok := cards[card]; ok {
			return nil, fmt.Errorf("duplicate card: '%s'", card)
		}
		cards[card] = struct{}{}
		deck = append(deck, protocol2.VoteValue(card))
	}

	return deck, nil
}

func runDeckAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		deck, err := parseDeck(args)
		if err != nil {
			return messages.NewErrorMessage(err)
		}

		err = m.app.Game.SetDeck(deck)
		return messages.NewErrorMessage(err)
	}
}

func runSelectAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("no issue index provided")
			return messages.NewErrorMessage(err)
		}

		index, err := strconv.Atoi(args[0])
		if err != nil {
			err = fmt.Errorf("invalid issue index: %s (%w)", args[0], err)
			return messages.NewErrorMessage(err)
		}

		return commands.SelectIssue(m.app, index)()
	}
}

func runRestoreAction(m *model, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			err := errors.New("no room id argument provided")
			return messages.NewErrorMessage(err)
		}

		roomID := protocol2.NewRoomID(args[0])
		state, err := m.app.LoadRoomState(roomID)
		if err != nil {
			return messages.NewErrorMessage(err)
		}
		if state == nil {
			err = errors.New("room not found in storage")
			return messages.NewErrorMessage(err)
		}
		return commands.JoinRoom(m.app, roomID, state)()
	}
}
