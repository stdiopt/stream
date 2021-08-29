package strmdiag

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	strm "github.com/stdiopt/stream"
)

// Counter provides a count diagnostics of message passed
type Counter struct {
	mu       sync.Mutex
	mark     time.Time
	duration time.Duration
	counts   map[string]int
	dcount   int

	lastCheck time.Time
	w         io.Writer
}

func NewCounter(w io.Writer, d time.Duration) *Counter {
	return &Counter{
		mark:     time.Now(),
		w:        w,
		duration: d,
	}
}

func (c *Counter) Add(v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.counts == nil {
		c.counts = map[string]int{}
	}
	now := time.Now()
	if now.After(c.mark.Add(c.duration)) {
		c.mark = now
		c.writeCount()
		c.dcount = 0
	}
	c.dcount++
	c.counts[fmt.Sprintf("%T", v)]++
}

func (c *Counter) WriteCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeCount()
}

func (c *Counter) writeCount() {
	dur := time.Since(c.lastCheck)
	c.lastCheck = time.Now()
	perSec := float64(c.dcount) / float64(dur) * float64(time.Second)
	buf := bytes.Buffer{}
	for k, v := range c.counts {
		fmt.Fprintf(&buf, "(%v: %s)", k, humanize(float64(v)))
	}
	fmt.Fprintf(c.w, "Processed messages: %v - %v/s\n", buf.String(), humanize(perSec))
}

func Count(d time.Duration) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		c := NewCounter(os.Stderr, d)
		defer c.WriteCount()
		return p.Consume(func(v interface{}) error {
			c.Add(v)
			return p.Send(v)
		})
	})
}

func humanize(count float64) string {
	units := []string{"", "k", "M", "T"}
	cur := 0
	for ; count > 1000 && cur < len(units); cur++ {
		count /= 1000
	}
	return fmt.Sprintf("%.02f%s", count, units[cur])
}
