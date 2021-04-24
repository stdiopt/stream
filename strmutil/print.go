package strmutil

import (
	"context"
	"fmt"
	"io"
)

func Print(w io.Writer, prefix string) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			fmt.Fprintf(w, "[%s] %v\n", prefix, v)
			return p.Send(ctx, v)
		})
	}
}
