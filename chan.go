package stream

import (
	"context"
	"fmt"
	"reflect"
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
func (c Chan) Consume(fn interface{}) error {
	cfn := MakeConsumerFunc(fn)
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case v, ok := <-c.ch:
			if !ok {
				return nil
			}
			if err := cfn(v.ctx, v.value); err != nil {
				return err
			}
		}
	}
}

// Close closes the channel.
func (c Chan) Close() {
	close(c.ch)
}

// Is a bit slower but some what wrappers typed messages.
func MakeConsumerFunc(fn interface{}) ConsumerFunc {
	if fn, ok := fn.(ConsumerFunc); ok {
		return fn
	}
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 2 {
		panic("should have 2 params")
	}
	args := make([]reflect.Value, 2)
	return func(ctx context.Context, v interface{}) error {
		args[0] = reflect.ValueOf(ctx)
		args[1] = reflect.ValueOf(v)
		if args[1].Type() != fnTyp.In(1) {
			return fmt.Errorf("invalid consumer type want: %v got %v", fnTyp.In(1), args[1].Type())
		}
		ret := fnVal.Call(args)
		if err, ok := ret[0].Interface().(error); ok && err != nil {
			return err
		}
		return nil
	}
}
