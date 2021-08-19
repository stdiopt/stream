package strmutil

import "github.com/stdiopt/stream"

func Pass(pass stream.P) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(pass.Send)
	})
}
