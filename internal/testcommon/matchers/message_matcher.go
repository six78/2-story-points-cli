package matchers

import (
	"2sp/pkg/protocol"
	"fmt"
)

type MessageMatcher struct {
	payload []byte
	message *protocol.Message
}

func (m *MessageMatcher) Matches(x interface{}) bool {
	m.message = nil
	m.payload = x.([]byte)
	if m.payload == nil {
		return false
	}

	message, err := protocol.UnmarshalMessage(m.payload)
	if err != nil {
		return false
	}

	m.message = message
	return true
}

func (m *MessageMatcher) String() string {
	return fmt.Sprintf("is protocol message")
}
