package stream

import (
	"context"
)

type ConsumerFunc = func(interface{}) error

// procChan wraps a channel and a context for cancellation awareness.
type procChan struct {
	ctx  context.Context
	ch   chan interface{}
	done chan struct{}
}

// newProcChan returns a Chan based on context with specific buffer size.
func newProcChan(ctx context.Context, buffer int) *procChan {
	return &procChan{
		ctx:  ctx,
		ch:   make(chan interface{}, buffer),
		done: make(chan struct{}),
	}
}

// Send sends v to the underlying channel if context is cancelled it will return
// the underlying ctx.Err()
func (c procChan) Send(v interface{}) error {
	// Check for done first
	select {
	case <-c.done:
		return ErrBreak
	default:
	}

	select {
	case <-c.done:
		return ErrBreak
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.ch <- v:
		return nil
	}
}

// Consume will consume a channel and call fn with the consumed value, it will
// block until either context is cancelled, channel is closed or ConsumerFunc
// error
// is not nil
func (c procChan) Consume(ifn interface{}) error {
	fn := MakeConsumerFunc(ifn)
	for {
		select {
		case <-c.done:
			return nil
		case <-c.ctx.Done():
			return c.ctx.Err()
		case v, ok := <-c.ch:
			if !ok {
				return nil
			}
			if err := fn(v); err != nil {
				return err
			}
		}
	}
}

func (c procChan) Cancel() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}

// Close closes the channel can close a closed channel.
func (c procChan) close() {
	close(c.ch)
}
