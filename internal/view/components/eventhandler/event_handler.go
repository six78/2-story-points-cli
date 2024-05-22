package eventhandler

import (
	tea "github.com/charmbracelet/bubbletea"
)

type buildMessageFunc[E any, M any] func(E) M

type subscription[E any, M any] struct {
	sub     chan E
	convert buildMessageFunc[E, M]
}

type Model[E any, M any] struct {
	// subscription is a pointer wrapper to share same subscription between models
	// and to enable nullifying it when channel is closed
	subscription *subscription[E, M]
}

func New[E any, M any](convert buildMessageFunc[E, M]) Model[E, M] {
	return Model[E, M]{
		subscription: &subscription[E, M]{
			sub:     nil,
			convert: convert,
		},
	}
}

func (m Model[E, M]) Init(input chan E, lastEvent E) tea.Cmd {
	m.subscription.sub = input
	m.subscription.sub <- lastEvent // Force notify current status after subscribing
	return WaitForEvent[E, M](m.subscription)
}

func (m Model[E, M]) Update(msg tea.Msg) (Model[E, M], tea.Cmd) {
	if m.subscription == nil || m.subscription.sub == nil {
		return m, nil
	}
	var cmd tea.Cmd
	switch msg.(type) {
	case M:
		cmd = WaitForEvent[E, M](m.subscription)
	}
	return m, cmd
}

func WaitForEvent[E any, M any](subscription *subscription[E, M]) tea.Cmd {
	return func() tea.Msg {
		if subscription.sub == nil {
			return nil
		}
		event, more := <-subscription.sub
		if more {
			return subscription.convert(event)
		}
		subscription.sub = nil
		return nil
	}
}
