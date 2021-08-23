package strmio

import (
	"github.com/stdiopt/stream"
	"golang.org/x/text/encoding/charmap"
)

// Might not be the proper package for this

// CharmapDecode receives bytes and converts from charmap to UTF-8.
func CharmapDecode(cm *charmap.Charmap) stream.Pipe {
	return stream.Func(func(p stream.Proc) error {
		dec := cm.NewDecoder()
		return p.Consume(func(buf []byte) error {
			res, err := dec.Bytes(buf)
			if err != nil {
				return err
			}
			return p.Send(res)
		})
	})
}
