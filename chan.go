package stream

import (
	"context"
)

// message is an internal message type that also passes ctx
type message struct {
	ctx   context.Context
	value interface{}
}

// Chan wraps a channel and a context for cancellation awareness.
type Chan struct {
	ctx context.Context
	ch  chan message
}

// NewChan returns a Chan based on context with specific buffer size.
func NewChan(ctx context.Context, buffer int) Chan {
	return Chan{
		ctx: ctx,
		ch:  make(chan message, buffer),
	}
}

// Send sends v to the underlying channel if context is cancelled it will return
// the underlying ctx.Err()
func (c Chan) Send(ctx context.Context, v interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.ch <- message{ctx, v}:
		return nil
	}
}

// Consume will consume a channel and call fn with the consumed value, it will
// block until either context is cancelled, channel is closed or ConsumerFunc
// error
// is not nil
func (c Chan) Consume(fn ConsumerFunc) error {
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case v, ok := <-c.ch:
			if !ok {
				return nil
			}
			if err := fn(v.ctx, v.value); err != nil {
				return err
			}
		}
	}
}

func (c Chan) Close() {
	close(c.ch)
}
