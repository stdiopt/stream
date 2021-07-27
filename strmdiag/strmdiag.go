package strmdiag

import (
	"fmt"
	"io"
	"time"

	"github.com/stdiopt/stream"
)

func Count(w io.Writer, d time.Duration) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		mark := time.Now()
		counts := map[string]int{}
		dcount := 0
		return p.Consume(func(v interface{}) error {
			now := time.Now()
			if now.After(mark.Add(d)) {
				mark = now
				perSec := float64(dcount) / float64(d) * float64(time.Second)
				fmt.Fprintf(w, "Processed messages: %v %.2f/s\n", counts, perSec)
				dcount = 0
			}
			dcount++
			counts[fmt.Sprintf("%T", v)]++
			return p.Send(v)
		})
	})
}

func Debug(w io.Writer) stream.Processor {
	return stream.F(func(p stream.Proc, v interface{}) error {
		stream.DebugProc(w, p, v)
		return p.Send(v)
	})
}
