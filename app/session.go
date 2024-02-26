package app

import (
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"strconv"
	pp "waku-poker-planning/protocol"
)

type Session struct {
	dealer       bool
	name         string
	contentTopic string
}

func NewSession(dealer bool, name string) (*Session, error) {
	contentTopic, err := calculateContentTopic(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate content topic")
	}
	return &Session{
		dealer:       dealer,
		name:         name,
		contentTopic: contentTopic.String(),
	}, nil
}

func calculateContentTopic(name string) (protocol.ContentTopic, error) {
	return protocol.NewContentTopic("six78", strconv.Itoa(pp.Version), name, "json")
}
