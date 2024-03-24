package waku

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"strconv"
	pp "waku-poker-planning/protocol"
)

type RoomCache struct {
	room         *pp.Room
	contentTopic string
	err          error
}

func NewRoomCache() RoomCache {
	return RoomCache{
		room:         nil,
		contentTopic: "",
	}
}

func (r *RoomCache) Get(room *pp.Room) (string, error) {
	if room != r.room {
		r.contentTopic, r.err = roomContentTopic(room)
	}
	return r.contentTopic, r.err
}

func roomContentTopic(info *pp.Room) (string, error) {
	if len(info.SymmetricKey) < 4 {
		return "", errors.New("symmetric key too short")
	}

	version := strconv.Itoa(int(pp.Version))
	contentTopicName := hexutil.Encode(info.SymmetricKey[:4])
	contentTopic, err := protocol.NewContentTopic("six78", version, contentTopicName, "json")

	if err != nil {
		return "", errors.Wrap(err, "failed to create content topic")
	}

	return contentTopic.String(), nil
}
