package protocol

import "time"

type VoteValue string

type VoteResult struct {
	Value     VoteValue `json:"estimation"` // TODO:  Vote -> Estimate ?
	Timestamp int64     `json:"timestamp"`
}

func NewVoteResult(value VoteValue) *VoteResult {
	return &VoteResult{
		Value:     value,
		Timestamp: time.Now().UnixMilli(),
	}
}

func (v *VoteResult) Hidden() VoteResult {
	return VoteResult{
		Value:     "",
		Timestamp: v.Timestamp,
	}
}
