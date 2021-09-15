package strmutil

import strm "github.com/stdiopt/stream"

// Value returns a ProcFunc that sends a single value v.
func Value(vs ...interface{}) strm.Pipe {
	return strm.S(func(s strm.Sender, _ interface{}) error {
		for _, v := range vs {
			if err := s.Send(v); err != nil {
				return err
			}
		}
		return nil
	})
}
