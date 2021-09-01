package stream

import (
	"context"
)

type ConsumerFunc = func(interface{}) error

// pipeChan wraps a channel and a context for cancellation awareness.
type pipeChan struct {
	// Maybe use a proc here?!
	ctx  context.Context
	ch   chan interface{}
	done chan struct{}
}

// newPipeChan returns a Chan based on context with specific buffer size.
func newPipeChan(ctx context.Context, buffer int) *pipeChan {
	return &pipeChan{
		ctx:  ctx,
		ch:   make(chan interface{}, buffer),
		done: make(chan struct{}),
	}
}

func (c pipeChan) Context() context.Context {
	return c.ctx
}

// Send sends v to the underlying channel if context is cancelled it will return
// the underlying ctx.Err()
func (c pipeChan) Send(v interface{}) error {
	// Check for canceled first then try to send
	// this way we avoid sending unexpectedly
	select {
	case <-c.done:
		return ErrBreak
	case <-c.ctx.Done():
		return c.ctx.Err()
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
func (c pipeChan) Consume(ifn interface{}) error {
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
			err := fn(v)
			if err == ErrBreak {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
}

func (c pipeChan) Println(...interface{}) {}

func (c pipeChan) Printf(string, ...interface{}) {}

func (c pipeChan) cancel() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}

// Close closes the channel can close a closed channel.
func (c pipeChan) close() {
	close(c.ch)
}
