package strmutil

import (
	"fmt"
	"io"
	"os"

	strm "github.com/stdiopt/stream"
)

var stdout = io.Writer(os.Stdout)

func SetOutput(w io.Writer) {
	stdout = w
}

// Print value with prefix.
func Print(prefix string) strm.Pipe {
	if prefix != "" {
		prefix = fmt.Sprintf("[%s] ", prefix)
	}
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(v interface{}) error {
			switch v := v.(type) {
			case []byte:
				fmt.Fprintf(stdout, "%s%v\n", prefix, string(v))
			default:
				fmt.Fprintf(stdout, "%s%v\n", prefix, v)
			}
			return p.Send(v)
		})
	})
}
