package strmutil

import "github.com/stdiopt/stream"

func Pass(pass stream.Sender) stream.Pipe {
	return stream.T(func(v interface{}) (interface{}, error) {
		if err := pass.Send(v); err != nil {
			return nil, err
		}
		return v, nil
	})
}
