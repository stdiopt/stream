package strmutil

import (
	"fmt"

	strm "github.com/stdiopt/stream"
)

// Repeat repeats last consumed value n times.
func Repeat(n int) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		if n <= 0 {
			return fmt.Errorf("invalid repeat param '%d', should be > 0", n)
		}
		return p.Consume(func(v interface{}) error {
			for i := 0; i < n; i++ {
				if err := p.Send(v); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
