package stream

import "io"

type DebugOpt struct {
	output    io.Writer
	Processor bool
	Value     bool
}

func Debug(w io.Writer) Processor {
	return Func(func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			p.Set("_debug", w)
			return p.Send(v)
		})
	})
}
