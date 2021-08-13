package strmutil

import (
	"github.com/stdiopt/stream"
)

func Limit(n int) stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		count := 0
		return p.Consume(func(v interface{}) error {
			if n == 0 {
				return stream.ErrBreak
			}

			if err := p.Send(v); err != nil {
				return err
			}
			count++
			if count >= n {
				return stream.ErrBreak
			}
			return nil
		})
	})
}
