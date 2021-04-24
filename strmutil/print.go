package strmutil

import (
	"fmt"
	"io"
)

func Print(w io.Writer, prefix string) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			fmt.Fprintf(w, "[%s] %v\n", prefix, v)
			return p.Send(v)
		})
	}
}
