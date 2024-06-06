package protocol

import (
	"time"
)

type Player struct {
	ID     PlayerID `json:"id"`
	Name   string   `json:"name"`
	Online bool     `json:"online"`

	// Deprecated: use OnlineTimestamp instead
	OnlineTimestamp             time.Time `json:"onlineTimestamp"`
	OnlineTimestampMilliseconds int64     `json:"onlineTimestampMilliseconds"`
}

func (p *Player) ApplyDeprecatedPatch() {
	if p.OnlineTimestampMilliseconds == 0 {
		p.OnlineTimestampMilliseconds = p.OnlineTimestamp.UnixMilli()
	}
}
