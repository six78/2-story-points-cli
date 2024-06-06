package protocol

import (
	"time"
)

type Player struct {
	ID     PlayerID `json:"id"`
	Name   string   `json:"name"`
	Online bool     `json:"online"`

	// Deprecated: use OnlineTimestamp instead
	// TODO: Those fields should be removed from json. They shouldn't be part of the protocol.
	// It should only be used by the dealer to keep the state of the player.
	// This becomes quickly inconsistent in the protocol, because when the dealer receives a new
	// player online message, he doesn't publish a new state with updated OnlineTimestamp. Because
	// that would increase the network usage unnecessarily.
	// Simple keeping this field in the struct, but removing it from json, works, but is not enough
	// for the dealer, because dealer wants to save users OnlineTimestamp in the storage.
	OnlineTimestamp             time.Time `json:"onlineTimestamp"`
	OnlineTimestampMilliseconds int64     `json:"onlineTimestampMilliseconds"`
}

func (p *Player) ApplyDeprecatedPatchOnReceive() {
	p.OnlineTimestampMilliseconds = p.OnlineTimestamp.UnixMilli()
}

func (p *Player) ApplyDeprecatedPatchOnSend() {
	p.OnlineTimestamp = time.UnixMilli(p.OnlineTimestampMilliseconds)
}

func (p *Player) OnlineTime() time.Time {
	return time.UnixMilli(p.OnlineTimestampMilliseconds)
}
