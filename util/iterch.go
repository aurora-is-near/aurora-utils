package util

import (
	"context"
	"iter"
)

type ChanIterator[T any] struct {
	ch     <-chan T
	closed bool
}

func NewChanIterator[T any](ch <-chan T) *ChanIterator[T] {
	return &ChanIterator[T]{
		ch:     ch,
		closed: false,
	}
}

func (it *ChanIterator[T]) Iterate(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-it.ch:
				if !ok {
					it.closed = true
					return
				}
				if !yield(v) {
					return
				}
			}
		}
	}
}

func (it *ChanIterator[T]) Closed() bool {
	return it.closed
}
