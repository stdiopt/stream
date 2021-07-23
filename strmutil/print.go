package strmutil

import (
	"context"
	"fmt"
	"io"

	"github.com/stdiopt/stream"
)

func Print(w io.Writer, prefix string) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			fmt.Fprintf(w, "[%s] %v\n", prefix, v)
			return p.Send(ctx, v)
		})
	})
}
