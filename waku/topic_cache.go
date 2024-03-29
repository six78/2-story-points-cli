package waku

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"go.uber.org/zap"
	"strconv"
	pp "waku-poker-planning/protocol"
)

type ContentTopicCache struct {
	logger       *zap.Logger
	room         *pp.Room
	contentTopic string
	err          error
}

func NewRoomCache(logger *zap.Logger) ContentTopicCache {
	return ContentTopicCache{
		logger:       logger.Named("TopicCache"),
		room:         nil,
		contentTopic: "",
	}
}

func (r *ContentTopicCache) Get(room *pp.Room) (string, error) {
	if room != r.room {
		r.contentTopic, r.err = r.roomContentTopic(room)
		if r.err != nil {
			r.logger.Error("failed to calculate content topic", zap.Error(r.err))
		} else {
			r.logger.Debug("new content topic", zap.String("contentTopic", r.contentTopic))
		}
	}
	return r.contentTopic, r.err
}

func (r *ContentTopicCache) roomContentTopic(room *pp.Room) (string, error) {
	roomID, err := room.ToRoomID()
	if err != nil {
		return "", errors.Wrap(err, "failed to create room ID")
	}

	version := strconv.Itoa(int(pp.Version))
	hash := crypto.Keccak256(room.Bytes())
	contentTopicName := hexutil.Encode(hash[:4])[2:]
	contentTopic, err := protocol.NewContentTopic("six78", version, contentTopicName, "json") // WARNING: "six78" is not the name of the app

	var bytesArray []string

	for _, b := range hash {
		bytesArray = append(bytesArray, strconv.Itoa(int(b)))
	}

	r.logger.Debug("content topic details",
		zap.String("version", version),
		zap.String("roomID", roomID.String()),
		zap.String("symmetricKey", hexutil.Encode(room.SymmetricKey)),
		zap.String("hashHexEncoded", hexutil.Encode(hash)),
		zap.Any("hashArray", bytesArray),
		zap.String("contentTopicName", contentTopicName),
		zap.String("contentTopic", contentTopic.String()),
	)

	if err != nil {
		return "", errors.Wrap(err, "failed to create content topic")
	}

	return contentTopic.String(), nil
}
