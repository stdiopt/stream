package strmutil

import (
	"fmt"

	strm "github.com/stdiopt/stream"
)

// Seq generates a sequence.
func Seq(start, end, step int) strm.Pipe {
	if (start == end) || (start > end && step >= 0) || (start < end && step <= 0) {
		panic(
			fmt.Sprintf("invalid range: %d->%d, step: %d causes infinite loop", start, end, step),
		)
	}
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(interface{}) error {
			if start > end {
				for i := start; i >= end; i += step {
					if err := p.Send(i); err != nil {
						return err
					}
				}
				return nil
			}
			for i := start; i <= end; i += step {
				if err := p.Send(i); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
