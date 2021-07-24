package strmutil

import (
	"time"

	"github.com/stdiopt/stream"
)

// End just a type used in several group streams to send the group
var End = struct{}{}

func Wait(d time.Duration) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			time.Sleep(d)
			return p.Send(v)
		})
	})
}
