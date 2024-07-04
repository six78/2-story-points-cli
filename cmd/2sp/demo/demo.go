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

var links = []string{
	"https://github.com/golang/go/issues/26492",
	"https://github.com/golang/go/issues/27605",
	"https://github.com/golang/go/issues/64997",
}
var votes = [][]string{
	{"3", "5", "3"},
	{"5", "8", "8"},
	{"13", "13", "8"},
}

type Demo struct {
	ctx     context.Context
	dealer  *game.Game
	state   game.StateSubscription
	program *tea.Program
	logger  *zap.Logger

	players    []*game.Game
	playerSubs []game.StateSubscription
}

func New(ctx context.Context, dealer *game.Game, program *tea.Program) *Demo {
	return &Demo{
		ctx:     ctx,
		dealer:  dealer,
		state:   dealer.SubscribeToStateChanges(),
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

func (d *Demo) initializePlayers() error {
	names := []string{"Alice", "Bob", "Charlie"}
	d.players = make([]*game.Game, len(names))
	d.playerSubs = make([]game.StateSubscription, len(names))

	wg := sync.WaitGroup{}
	errChan := make(chan error, len(names))

	for i, name := range names {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			player, err := d.createPlayer(name)
			if err != nil {
				errChan <- errors.Wrap(err, "failed to create player")
				return
			}
			d.players[i] = player
			d.playerSubs[i] = player.SubscribeToStateChanges()
		}(i, name)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Demo) Routine() {
	defer d.Stop()

	d.logger.Info("started")
	time.Sleep(2 * time.Second) // Wait for the program to start

	// Create new room
	d.sendShortcut(commands.DefaultKeyMap.NewRoom)
	d.logger.Info("room created")

	// Add players
	err := d.initializePlayers()
	if err != nil {
		d.logger.Error("failed to initialize players", zap.Error(err))
		return
	}

	err = d.waitForPlayers(d.players)
	if err != nil {
		d.logger.Error("failed to wait for players", zap.Error(err))
		return
	}
	d.logger.Info("players joined")

	// Switch to issues view
	d.sendKey(tea.KeyTab, 500*time.Millisecond)
	d.logger.Info("switched to issues view")

	// Add issues
	d.sendText(strings.Join(links, "\n"))
	d.logger.Info("issues added")
	time.Sleep(500 * time.Millisecond)

	// Wait for issues to be added
	err = d.waitForIssues(len(links))
	if err != nil {
		d.logger.Error("failed to wait for issues", zap.Error(err))
		return
	}

	// Players vote for issues
	issues := d.dealer.CurrentState().Issues
	for i, issue := range issues {
		err = d.issueSubRoutine(issue.ID, votes[i])
		if err != nil {
			d.logger.Error("failed to run issue sub-routine", zap.Error(err),
				zap.String("issueLink", issue.TitleOrURL))
			return
		}
	}
	time.Sleep(2000 * time.Millisecond)

	// Switch to issues view
	d.sendKey(tea.KeyTab, 500*time.Millisecond)
	d.logger.Info("switched to issues view")

	d.sendKey(tea.KeyDown, 200*time.Millisecond)
	d.sendKey(tea.KeyDown, 1000*time.Millisecond)

	d.sendKey(tea.KeyUp, 200*time.Millisecond)
	d.sendKey(tea.KeyUp, 3*time.Second)

	d.logger.Info("finished")
}

func (d *Demo) issueSubRoutine(issueID protocol.IssueID, votes []string) error {
	errChan := make(chan error, len(votes))
	wg := sync.WaitGroup{}
	wg.Add(len(votes))
	for i, vote := range votes {
		go func(index int, vote string) {
			defer wg.Done()
			errChan <- d.playerVoteSubroutine(index, issueID, protocol.VoteValue(vote))
		}(i, vote)
	}

	// Deal first issue (expect all players to be subscribed to state changes
	d.sendKey(tea.KeyEnter, 1*time.Second)
	d.logger.Info("deal issue")

	// Dealer vote with UI
	d.sendKey(tea.KeyRight, 500*time.Millisecond)
	d.sendKey(tea.KeyRight, 500*time.Millisecond)
	d.sendKey(tea.KeyEnter, 0)

	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	d.logger.Info("players voted")

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
	d.sendKey(tea.KeyEnter, 0)

	return nil
}

func (d *Demo) playerVoteSubroutine(index int, issueID protocol.IssueID, vote protocol.VoteValue) error {
	player := d.players[index]
	playerName := player.Player().Name

	err := d.waitForIssueDealt(d.playerSubs[index], issueID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%s: issue not dealt", playerName))
	}

	// Random delay to simulate human behavior
	delay := time.Duration(rand.Intn(4000)) * time.Millisecond
	time.Sleep(delay)

	// Publish vote
	err = player.PublishVote(vote)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%s: failed to publish vote", playerName))
	}

	return nil
}

func (d *Demo) sendShortcut(key key.Binding) {
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(key.Keys()[0]),
	}
	d.program.Send(keyMsg)
}

func (d *Demo) sendKey(key tea.KeyType, sleepDuration time.Duration) {
	keyMsg := tea.KeyMsg{
		Type: key,
	}
	d.program.Send(keyMsg)
	time.Sleep(sleepDuration)
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
