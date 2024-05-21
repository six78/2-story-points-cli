package transporteventhandler

import (
	"2sp/internal/transport"
	"2sp/internal/view/messages"
	tea "github.com/charmbracelet/bubbletea"
)

// NOTE: This component is currently implemented with no support of multiple games.
// 		 The assumption is that only one object sends `GameStateMessage`s.
//		 generate could be a solution here.

// NOTE: This component is very similar to GameEventHandler. But I decided to keep them separate for now.
// 		 This can be refactored later when both components are stable.

type subscription struct {
	sub transport.ConnectionStatusSubscription
}

type Model struct {
	transport transport.Service

	// subscription is a pointer wrapper to share same subscription between models
	// and to enable nullifying it when channel is closed
	subscription *subscription
}

func New(transport transport.Service) Model {
	return Model{
		transport: transport,
		subscription: &subscription{
			sub: nil,
		},
	}
}

func (m Model) Init() tea.Cmd {
	if m.transport == nil {
		return nil
	}
	m.subscription.sub = m.transport.SubscribeToConnectionStatus()
	m.subscription.sub <- m.transport.ConnectionStatus() // Force notify current status after subscribing
	return WaitForConnectionStatus(m.subscription)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.subscription == nil || m.subscription.sub == nil {
		return m, nil
	}
	var cmd tea.Cmd
	switch msg.(type) {
	case messages.ConnectionStatus:
		cmd = WaitForConnectionStatus(m.subscription)
	}
	return m, cmd
}

func WaitForConnectionStatus(subscription *subscription) tea.Cmd {
	return func() tea.Msg {
		if subscription.sub == nil {
			return nil
		}
		status, more := <-subscription.sub
		if more {
			return messages.ConnectionStatus{
				Status: status,
			}
		}
		subscription.sub = nil
		return nil
	}
}
