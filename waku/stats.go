package waku

type Stats struct {
	sentMessages     int
	receivedMessages int
}

func NewStats() *Stats {
	return &Stats{
		sentMessages:     0,
		receivedMessages: 0,
	}
}

func (s *Stats) IncrementSentMessages() {
	s.sentMessages++
}

func (s *Stats) IncrementReceivedMessages() {
	s.receivedMessages++
}

func (s *Stats) SentMessages() int {
	return s.sentMessages
}

func (s *Stats) ReceivedMessages() int {
	return s.receivedMessages
}
