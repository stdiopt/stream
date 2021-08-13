package strmutil

import (
	"fmt"
	"io"

	"github.com/stdiopt/stream"
)

func Print(prefix string, w io.Writer) stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			fmt.Fprintf(w, "[%s] %v\n", prefix, v)
			return p.Send(v)
		})
	})
}
