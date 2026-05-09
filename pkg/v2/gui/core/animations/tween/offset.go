package tween

import "sync/atomic"

type Offset struct {
	offset         atomic.Int64
	startingOffset atomic.Int64
	endingOffset   atomic.Int64
}

func (o *Offset) Current() int {
	return int(o.offset.Load())
}

func (o *Offset) Start() int {
	return int(o.startingOffset.Load())
}

func (o *Offset) End() int {
	return int(o.endingOffset.Load())
}
