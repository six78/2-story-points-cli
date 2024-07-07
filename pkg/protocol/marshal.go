package protocol

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func UnmarshalMessage(payload []byte) (*Message, error) {
	message := Message{}
	err := json.Unmarshal(payload, &message)
	return &message, errors.Wrap(err, "failed to unmarshal message")
}

func UnmarshalPlayerVote(payload []byte) (*PlayerVoteMessage, error) {
	vote := PlayerVoteMessage{}

	err := json.Unmarshal(payload, &vote)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal message")
	}

	if vote.Type != MessageTypePlayerVote {
		return nil, errors.New("message is not a player vote message")
	}

	return &vote, errors.Wrap(err, "failed to unmarshal player vote message")
}
