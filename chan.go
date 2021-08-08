package stream

import (
	"context"
)

// procChan wraps a channel and a context for cancellation awareness.
type procChan struct {
	ch   chan Message
	done chan struct{}
}

// newProcChan returns a Chan based on context with specific buffer size.
func newProcChan(buffer int) *procChan {
	return &procChan{
		ch:   make(chan Message, buffer),
		done: make(chan struct{}),
	}
}

// Send sends v to the underlying channel if context is cancelled it will return
// the underlying ctx.Err()
func (c procChan) send(ctx context.Context, m Message) error {
	select {
	case <-c.done:
		return ErrBreak
	case <-ctx.Done():
		return ctx.Err()
	case c.ch <- m:
		return nil
	}
}

// Consume will consume a channel and call fn with the consumed value, it will
// block until either context is cancelled, channel is closed or ConsumerFunc
// error
// is not nil
func (c procChan) consume(ctx context.Context, fn func(Message) error) error {
	for {
		select {
		case <-c.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-c.ch:
			if !ok {
				return nil
			}
			err := fn(m)
			if err == ErrBreak {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
}

func (c procChan) cancel() {
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
