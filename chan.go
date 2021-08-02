package stream

import (
	"context"
	"sync"
)

// procChan wraps a channel and a context for cancellation awareness.
type procChan struct {
	ctx        context.Context
	ch         chan Message
	done       chan struct{} // we can close an individual channel
	cancelOnce sync.Once
}

// newProcChan returns a Chan based on context with specific buffer size.
func newProcChan(ctx context.Context, buffer int) *procChan {
	return &procChan{
		ctx:  ctx,
		ch:   make(chan Message, buffer),
		done: make(chan struct{}),
	}
}

// Send sends v to the underlying channel if context is cancelled it will return
// the underlying ctx.Err()
func (c procChan) send(m Message) error {
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
func (c procChan) consume(fn func(Message) error) error {
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

func (c *procChan) cancel() {
	// c.cancelOnce.Do(func() { close(c.done) })
	close(c.done)
}

// Close closes the channel.
func (c procChan) close() {
	close(c.ch)
}
