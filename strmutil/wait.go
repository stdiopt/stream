package strmutil

import (
	"time"

	strm "github.com/stdiopt/stream"
)

func Wait(d time.Duration) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(v interface{}) error {
			time.Sleep(d)
			return p.Send(v)
		})
	})
}
