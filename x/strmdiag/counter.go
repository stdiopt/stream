package strmdiag

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	strm "github.com/stdiopt/stream"
)

func Count(label string, d time.Duration) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var (
			count     uint64
			lastCount uint64
			lastCheck time.Time
		)
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		lastCheck = time.Now()
		go func() {
			for range ticker.C {
				dur := time.Since(lastCheck)
				lastCheck = time.Now()
				cc := atomic.LoadUint64(&count)
				dcount := cc - lastCount
				lastCount = cc
				perSec := float64(dcount) / float64(dur) * float64(time.Second)
				fmt.Fprintf(os.Stderr, "[%s] Processed messages: %v - %v/s\n", label, humanize(float64(count)), humanize(perSec))
			}
		}()
		return p.Consume(func(v interface{}) error {
			atomic.AddUint64(&count, 1)
			return p.Send(v)
		})
	})
}
