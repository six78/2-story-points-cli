package gameeventhandler

import (
	"2sp/internal/config"
	"2sp/internal/view/messages"
	"2sp/pkg/game"
	tea "github.com/charmbracelet/bubbletea"
)

// NOTE: This component is currently implemented with no support of multiple games.
// 		 The assumption is that only one object sends `GameStateMessage`s.

type subscription struct {
	sub game.StateSubscription
}

type Model struct {
	game *game.Game

	// subscription is a pointer wrapper to share same subscription between models
	// and to enable nullifying it when channel is closed
	subscription *subscription
}

func New(game *game.Game) Model {
	return Model{
		game: game,
		subscription: &subscription{
			sub: nil,
		},
	}
}

func (m Model) Init() tea.Cmd {
	if m.game == nil {
		return nil
	}
	m.subscription.sub = m.game.SubscribeToStateChanges()
	return WaitForGameState(m.subscription)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.subscription == nil || m.subscription.sub == nil {
		return m, nil
	}
	var cmd tea.Cmd
	switch msg.(type) {
	case messages.GameStateMessage:
		cmd = WaitForGameState(m.subscription)
	}
	return m, cmd
}

func WaitForGameState(subscription *subscription) tea.Cmd {
	return func() tea.Msg {
		if subscription.sub == nil {
			config.Logger.Error("game state subscription is not created")
			return nil
		}
		state, more := <-subscription.sub
		if more {
			return messages.GameStateMessage{State: state}
		}
		subscription.sub = nil
		return nil
	}
}
