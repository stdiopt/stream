package strmutil

import strm "github.com/stdiopt/stream"

// Value returns a ProcFunc that sends a single value v.
func Value(vs ...interface{}) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		p.Consume(func(interface{}) error {
			for _, v := range vs {
				if err := p.Send(v); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
