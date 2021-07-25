package stream

import (
	"context"
)

// message is an internal message type that also passes ctx
type message struct {
	value interface{}
	meta  map[string]interface{}
}

// Chan wraps a channel and a context for cancellation awareness.
type Chan struct {
	ctx  context.Context
	ch   chan message
	done chan struct{} // we can close an individual channel
}

// newChan returns a Chan based on context with specific buffer size.
func newChan(ctx context.Context, buffer int) Chan {
	return Chan{
		ctx:  ctx,
		ch:   make(chan message, buffer),
		done: make(chan struct{}),
	}
}

// Send sends v to the underlying channel if context is cancelled it will return
// the underlying ctx.Err()
func (c Chan) send(m message) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-c.done:
		return ErrBreak
	case c.ch <- m:
		return nil
	}
}

// Consume will consume a channel and call fn with the consumed value, it will
// block until either context is cancelled, channel is closed or ConsumerFunc
// error
// is not nil
func (c Chan) consume(fn func(message) error) error {
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-c.done:
			return nil
		case m, ok := <-c.ch:
			if !ok {
				return nil
			}
			if err := fn(m); err != nil {
				return err
			}
		}
	}
}

func (c Chan) cancel() {
	close(c.done)
}

// Close closes the channel.
func (c Chan) Close() {
	close(c.ch)
}
