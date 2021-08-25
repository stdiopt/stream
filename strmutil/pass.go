package strmutil

import strm "github.com/stdiopt/stream"

// Pass pass value to another sender, usefull for nested pipelines.
func Pass(pass strm.Sender) strm.Pipe {
	return strm.T(func(v interface{}) (interface{}, error) {
		if err := pass.Send(v); err != nil {
			return nil, err
		}
		return v, nil
	})
}
