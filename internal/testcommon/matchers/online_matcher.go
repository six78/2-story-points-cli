package matchers

import (
	"2sp/pkg/protocol"
	"fmt"
)

type OnlineMatcher struct {
	MessageMatcher
}

func NewOnlineMatcher() OnlineMatcher {
	return OnlineMatcher{}
}

func (m OnlineMatcher) Matches(x interface{}) bool {
	if !m.MessageMatcher.Matches(x) {
		return false
	}

	if m.message.Type != protocol.MessageTypePlayerOnline {
		return false
	}

	return true
}

func (m OnlineMatcher) String() string {
	return fmt.Sprintf("is user online protocol message")
}
