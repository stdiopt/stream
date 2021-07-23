package strmutil

import (
	"context"
	"io"

	"github.com/stdiopt/stream"
)

func Debug(w io.Writer) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			ctx = stream.ContextWithMetaValue(ctx, "_debug", w)
			return p.Send(ctx, v)
		})
	})
}
