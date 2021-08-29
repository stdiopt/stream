package strmutil

import (
	"time"

	strm "github.com/stdiopt/stream"
)

func Wait(d time.Duration) strm.Pipe {
	return strm.S(func(s strm.Sender, v interface{}) error {
		time.Sleep(d)
		return s.Send(v)
	})
}
