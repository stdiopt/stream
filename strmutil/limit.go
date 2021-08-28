package strmutil

import (
	"io"

	strm "github.com/stdiopt/stream"
)

// Limit passes N values and breaks the pipeline.
func Limit(n int) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		count := 0
		return p.Consume(func(v interface{}) error {
			if n == 0 {
				return io.EOF
			}

			if err := p.Send(v); err != nil {
				return err
			}
			count++
			if count >= n {
				return io.EOF
			}
			return nil
		})
	})
}
