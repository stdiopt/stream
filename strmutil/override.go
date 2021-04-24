package strmutil

import (
	"context"

	"github.com/stdiopt/stream"
)

type ProcOverride struct {
	Proc
	ConsumerFunc func(stream.ConsumerFunc) error
	SenderFunc   func(context.Context, interface{}) error
}

func (w ProcOverride) Consume(fn stream.ConsumerFunc) error {
	if w.ConsumerFunc != nil {
		return w.ConsumerFunc(fn)
	}
	return w.Proc.Consume(fn)
}

func (w ProcOverride) Send(ctx context.Context, v interface{}) error {
	if w.SenderFunc != nil {
		return w.SenderFunc(ctx, v)
	}
	return w.Proc.Send(ctx, v)
}
