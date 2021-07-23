package strmutil

import (
	"context"

	"github.com/stdiopt/stream"
)

// Value returns a ProcFunc that sends a single value v.
func Value(v interface{}) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Send(p.Context(), v)
	})
}

// Repeat consumes and sends n times.
func Repeat(n int) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			for i := 0; i < n; i++ {
				if err := p.Send(ctx, v); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
