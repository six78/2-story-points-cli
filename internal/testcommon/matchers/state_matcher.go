package matchers

import (
	"encoding/json"
	"testing"

	"github.com/six78/2-story-points-cli/pkg/protocol"
)

type Callback func(state *protocol.State) bool

type StateMatcher struct {
	Matcher
	MessageMatcher
	cb    Callback
	state protocol.State
}

func NewStateMatcher(t *testing.T, cb Callback) *StateMatcher {
	return &StateMatcher{
		Matcher: *NewMatcher(t),
		cb:      cb,
	}
}

func (m *StateMatcher) Matches(x interface{}) bool {
	var state protocol.State

	switch x := x.(type) {
	case []byte:
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
		state = stateMessage.State

	case *protocol.State:
		state = *x

	case protocol.State:
		state = x

	default:
		return false
	}

	m.state = state
	m.triggered <- state

	if m.cb == nil {
		return true
	}

	return m.cb(&state)
}

func (m *StateMatcher) String() string {
	return "is state message matching custom condition"
}

func (m *StateMatcher) State() protocol.State {
	return m.state
}

func (m *StateMatcher) Wait() protocol.State {
	return m.Matcher.Wait().(protocol.State)
}
