package demo

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/six78/2-story-points-cli/internal/config"
	"github.com/six78/2-story-points-cli/internal/transport"
	"github.com/six78/2-story-points-cli/internal/view/commands"
	"github.com/six78/2-story-points-cli/pkg/game"
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

type Demo struct {
	ctx     context.Context
	dealer  *game.Game
	state   game.StateSubscription
	program *tea.Program
	logger  *zap.Logger

	players    []*game.Game
	playerSubs []game.StateSubscription
}

func New(ctx context.Context, game *game.Game, program *tea.Program) *Demo {
	return &Demo{
		ctx:     ctx,
		dealer:  game,
		state:   game.SubscribeToStateChanges(),
		program: program,
		logger:  config.Logger.Named("demo"),
	}
}

func (d *Demo) Stop() {
	d.logger.Info("stopping")

	for _, player := range d.players {
		player.Stop()
	}
}

func (d *Demo) Routine() {
	defer d.Stop()

	d.logger.Info("started")

	// TODO: wait for the program to start
	time.Sleep(2 * time.Second)
	d.logger.Info("initial wait finished")

	// Create new room
	d.sendShortcut(commands.DefaultKeyMap.NewRoom)
	d.logger.Info("room created")

	// Add players
	playersNames := []string{"Alice", "Bob", "Charlie"}
	d.players = make([]*game.Game, 0, 3)
	d.playerSubs = make([]game.StateSubscription, 0, 3)

	for _, name := range playersNames {
		player, err := d.createPlayer(name)
		if err != nil {
			d.logger.Error("failed to create player", zap.Error(err))
			return
		}
		d.players = append(d.players, player)
		d.playerSubs = append(d.playerSubs, player.SubscribeToStateChanges())
	}

	err := d.waitForPlayers(d.players)
	if err != nil {
		d.logger.Error("failed to wait for players", zap.Error(err))
		return
	}
	d.logger.Info("players joined")

	// Switch to issues view
	d.sendKey(tea.KeyTab)
	d.logger.Info("switched to issues view")
	time.Sleep(500 * time.Millisecond)

	// Add issues
	links := []string{
		"https://github.com/golang/go/issues/26492",
		"https://github.com/golang/go/issues/27605",
		"https://github.com/golang/go/issues/64997",
	}
	votes := [][]string{
		{"3", "5", "3"},
		{"5", "8", "8"},
		{"13", "13", "8"},
	}
	//issuesLinks := maps.Keys(issues)

	d.sendText(strings.Join(links, "\n"))
	d.logger.Info("issues added")
	time.Sleep(500 * time.Millisecond)

	// Wait for issues to be added
	err = d.waitForIssues(len(links))
	if err != nil {
		d.logger.Error("failed to wait for issues", zap.Error(err))
		return
	}

	issues := d.dealer.CurrentState().Issues
	for i, issue := range issues {
		err = d.issueSubRoutine(issue.ID, votes[i])
		if err != nil {
			d.logger.Error("failed to run issue sub-routine",
				zap.Error(err),
				zap.String("issueLink", issue.TitleOrURL))
			return
		}
	}
	time.Sleep(2000 * time.Millisecond)

	// Switch to issues view
	d.sendKey(tea.KeyTab)
	d.logger.Info("switched to issues view")
	time.Sleep(500 * time.Millisecond)

	d.sendKey(tea.KeyDown)
	time.Sleep(200 * time.Millisecond)
	d.sendKey(tea.KeyDown)
	time.Sleep(1000 * time.Millisecond)

	d.sendKey(tea.KeyUp)
	time.Sleep(200 * time.Millisecond)
	d.sendKey(tea.KeyUp)
	time.Sleep(3 * time.Second)

	d.logger.Info("finished")
}

func (d *Demo) issueSubRoutine(issueID protocol.IssueID, votes []string) error {
	var errs []error
	wg := sync.WaitGroup{}
	wg.Add(len(votes))
	for i, vote := range votes {
		go func(i int, vote string) {
			defer wg.Done()
			player := d.players[i]
			playerName := player.Player().Name

			err := d.waitForIssueDealt(d.playerSubs[i], issueID)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("%s: issue not dealt", playerName))
				errs = append(errs, err)
				return
			}

			// Random delay to simulate human behavior
			delay := time.Duration(rand.Intn(4000)) * time.Millisecond
			time.Sleep(delay)

			err = player.PublishVote(protocol.VoteValue(vote))
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("%s: failed to publish vote", playerName))
				errs = append(errs, err)
				return
			}
		}(i, vote)
	}

	// Deal first issue (expect all players to be subscribed to state changes
	d.sendKey(tea.KeyEnter)
	d.logger.Info("deal issue")

	// TODO: dealer vote by arrows
	time.Sleep(1000 * time.Millisecond)
	d.sendKey(tea.KeyRight)
	time.Sleep(500 * time.Millisecond)
	d.sendKey(tea.KeyRight)
	time.Sleep(500 * time.Millisecond)
	d.sendKey(tea.KeyEnter)

	wg.Wait()
	d.logger.Info("players voted")

	if len(errs) > 0 {
		d.logger.Error("errors occurred", zap.Errors("errors", errs))
		return errors.New("failed to publish votes")
	}

	err := d.waitForVotes(votes)
	if err != nil {
		err = errors.Wrap(err, "failed to wait for votes")
		return err
	}
	time.Sleep(1 * time.Second)

	// Reveal votes
	d.sendShortcut(commands.DefaultKeyMap.RevealVotes)
	d.logger.Info("votes revealed")
	err = d.waitForStateCondition(d.state, func(state *protocol.State) bool {
		return state.VotesRevealed
	})
	if err != nil {
		err = errors.Wrap(err, "failed to wait for votes to be revealed")
		return err
	}
	time.Sleep(2 * time.Second)

	// Finish vote
	d.sendKey(tea.KeyEnter)

	return nil
}

func (d *Demo) sendShortcut(key key.Binding) {
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(key.Keys()[0]),
	}
	d.program.Send(keyMsg)
}

func (d *Demo) sendKey(key tea.KeyType) {
	keyMsg := tea.KeyMsg{
		Type: key,
	}
	d.program.Send(keyMsg)
}

func (d *Demo) sendText(text string) {
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(text),
	}
	d.program.Send(keyMsg)
}

func (d *Demo) createPlayer(name string) (*game.Game, error) {
	logger := config.Logger.Named(strings.ToLower(name))

	tr := transport.NewNode(d.ctx, logger)
	err := tr.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize transport")
	}

	err = tr.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transport")
	}

	player := game.NewGame([]game.Option{
		game.WithContext(d.ctx),
		game.WithTransport(tr),
		game.WithPlayerName(name),
		game.WithClock(clockwork.NewRealClock()),
		game.WithLogger(logger),
	})

	err = player.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize dealer")
	}

	err = player.JoinRoom(d.dealer.RoomID(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to join room")
	}

	return player, nil
}

func (d *Demo) waitForStateCondition(sub game.StateSubscription, condition func(state *protocol.State) bool) error {
	timeout := time.After(10 * time.Second)
	for {
		select {
		case state := <-sub:
			if condition(state) {
				time.Sleep(500 * time.Millisecond)
				return nil
			}
		case <-timeout:
			return errors.New("timeout waiting for state condition")
		case <-d.ctx.Done():
		}
	}
}

func (d *Demo) waitForPlayers(players []*game.Game) error {
	return d.waitForStateCondition(d.state, func(state *protocol.State) bool {
		return len(state.Players) == len(players)
	})
}

func (d *Demo) waitForIssues(count int) error {
	return d.waitForStateCondition(d.state, func(state *protocol.State) bool {
		return len(state.Issues) == count
	})
}

func (d *Demo) waitForVotes(votes []string) error {
	return d.waitForStateCondition(d.state, func(state *protocol.State) bool {
		issue := state.GetActiveIssue()
		if issue == nil {
			return false
		}
		return len(issue.Votes) == len(votes)
	})
}

func (d *Demo) waitForIssueDealt(sub game.StateSubscription, issueID protocol.IssueID) error {
	return d.waitForStateCondition(sub, func(state *protocol.State) bool {
		return state.ActiveIssue == issueID
	})
}
