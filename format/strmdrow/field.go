package strmdrow

import (
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

func Field(k string) strm.Pipe {
	return strm.S(func(s strm.Sender, row drow.Row) error {
		return s.Send(row.Get(k))
	})
}
