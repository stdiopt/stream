package strmutil

import strm "github.com/stdiopt/stream"

// Pass pass value to another sender, usefull for nested pipelines.
func Pass(pass strm.Sender) strm.Pipe {
	return strm.S(func(s strm.Sender, v interface{}) error {
		if err := pass.Send(v); err != nil {
			return err
		}
		return s.Send(v)
	})
}
