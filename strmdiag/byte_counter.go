package strmdiag

import (
	"fmt"
	"os"
	"time"

	strm "github.com/stdiopt/stream"
)

func ByteCount(label string, d time.Duration) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var (
			count     int
			lastCount int
			lastCheck time.Time
		)
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				dur := time.Since(lastCheck)
				lastCheck = time.Now()
				dcount := count - lastCount
				lastCount = count
				perSec := float64(dcount) / float64(dur) * float64(time.Second)
				fmt.Fprintf(os.Stderr, "[%s] bytes processed: %v - %v/s\n", label, humanize(float64(count)), humanize(perSec))
			}
		}()
		return p.Consume(func(b []byte) error {
			count += len(b)
			return p.Send(b)
		})
	})
}
