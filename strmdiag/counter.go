package strmdiag

import (
	"os"
	"time"

	strm "github.com/stdiopt/stream"
)

func Count(label string, d time.Duration) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		metrics := NewMetrics(os.Stderr, d)
		count := metrics.Count(label)
		count.Run(p.Context(), p, p)
	})
}
