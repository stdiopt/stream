package strmutil

import strm "github.com/stdiopt/stream"

// Pass pass value to another sender, usefull for nested pipelines.
func Pass(pass strm.Sender) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(v interface{}) error {
			if err := pass.Send(v); err != nil {
				return err
			}
			return p.Send(v)
		})
	})
}
