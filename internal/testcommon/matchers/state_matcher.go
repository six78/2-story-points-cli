package matchers

import (
	"2sp/pkg/protocol"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type Matcher func(state *protocol.State)

type StateMatcher struct {
	MessageMatcher
	matcher Matcher
	state   protocol.State

	triggered chan protocol.State
}

func NewStateMatcher(matcher Matcher) *StateMatcher {
	return &StateMatcher{
		matcher:   matcher,
		triggered: make(chan protocol.State, 1),
	}
}

func (m *StateMatcher) Matches(x interface{}) bool {
	defer func() {
		m.triggered <- m.state
	}()

	if !m.MessageMatcher.Matches(x) {
		return false
	}

	if m.message.Type != protocol.MessageTypeState {
		return false
	}

	var stateMessage protocol.GameStateMessage
	err := json.Unmarshal(m.payload, &stateMessage)
	if err != nil {
		return false
	}

	m.state = stateMessage.State

	if m.matcher == nil {
		return true
	}

	m.matcher(&stateMessage.State)
	return true
}

func (m *StateMatcher) String() string {
	return fmt.Sprintf("is state message matching custom condition")
}

func (m *StateMatcher) State() protocol.State {
	return m.state
}

func (m *StateMatcher) Wait(t *testing.T) protocol.State {
	select {
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for state message")
	case <-m.triggered:
	}
	return m.state
}
