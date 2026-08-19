package grpc

import "sync/atomic"

type ISequencer interface {
	Next() uint64
}

type AtomicSequencer struct {
	next atomic.Uint64
}

func NewAtomicSequencer() *AtomicSequencer {
	return &AtomicSequencer{
		next: atomic.Uint64{},
	}
}

func (s *AtomicSequencer) Next() uint64 {
	return s.next.Add(1)
}
